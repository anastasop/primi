package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"image"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/bmizerany/pat"
)

const (
	// keyExpirationInMinutesall is the duration after which
	// all in-memory images are discarded
	keyExpirationInMinutes = time.Duration(5) * time.Minute
)

// downloadAndTransformImage does an HTTP GET for rawurl
// and if it is a jpeg or png image it applies the transform
func downloadAndTransformImage(rawurl string, transform func(image.Image) image.Image) (image.Image, error) {
	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawimg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(rawimg))
	if err != nil {
		return nil, err
	}

	return transform(img), nil
}

// transformRequest is the payload of an HTTP request for an image transformation
type transformRequest struct {
	rawurl string
	draw   string
}

// transformRequest is the result of an image transformation
// Message is a human readable message for the user
// Image is the actual transformed image
// GetURL is the url for the transformed image
// A json serialization of transformResult is return to the browser
// as a server side event
type transformResult struct {
	Message string `json:"message"`
	Image   []byte `json:"-"`
	GetURL  string `json:"url"`
	ImgURL  string `json:"img"`
	SrcURL  string `json:"src"`
}

// transforms is an in-memory storage for transformed images
var transforms struct {
	sync.Mutex
	images map[string]*transformResult
}

// transformer receives a request from in, executes it and responds
// to out. The requests come from a separate HTTP handler and the
// responses always go to the sse endpoint
func transformer(in <-chan *transformRequest, out chan<- *event) {
	for {
		req := <-in
		log.Printf("transformer starts for %q and %q", req.draw, req.rawurl)

		// beware of shadowing err. Statuses are propagated to
		// the final error handler
		// TODO fix it
		var img image.Image
		var err error
		switch req.draw {
		case "primitive":
			img, err = downloadAndTransformImage(req.rawurl, primitiveOnImage)
		case "triangle":
			img, err = downloadAndTransformImage(req.rawurl, triangleOnImage)
		}
		if err == nil {
			var b bytes.Buffer
			if err = png.Encode(&b, img); err == nil {
				key := fmt.Sprintf("%x", sha1.Sum(b.Bytes()))[0:8]
				res := &transformResult{
					Message: fmt.Sprintf("Ready: %s", req.rawurl),
					Image:   b.Bytes(),
					GetURL:  fmt.Sprintf("/show/%s", key),
					ImgURL:  fmt.Sprintf("/image/%s", key),
					SrcURL:  req.rawurl,
				}

				var enc []byte
				enc, err = json.Marshal(res)
				if err == nil {
					transforms.Lock()
					transforms.images[key] = res
					transforms.Unlock()
					// easier to start a goroutine for expiration
					// than implement a form of GC with timestamps in transforms
					go func(k, u string) {
						<-time.After(keyExpirationInMinutes)
						transforms.Lock()
						delete(transforms.images, k)
						transforms.Unlock()
						log.Printf("expired key %s for %q", k, u)
					}(key, req.rawurl)

					out <- &event{"image", string(enc), ""}
					log.Printf("transformer finished for %q and %q", req.draw, req.rawurl)
				}
			}
		}
		if err != nil {
			log.Printf("transformer failed for %q: %s", req.rawurl, err.Error())
			out <- &event{"problem", err.Error(), ""}
		}
	}
}

// event is a SSE event
type event struct {
	Event       string
	Data        string
	LastEventID string
}

// newSSEHandler is an HTTP handler for server side events
// each value for ec becomes an sse event
// also every d sends a comment to keep the connection alive
func newSSEHandler(ec chan *event, d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("accepted connection from", req.RemoteAddr)

		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "ResponseWriter does not support Flusher", http.StatusInternalServerError)
			return
		}

		cn, ok := w.(http.CloseNotifier)
		if !ok {
			http.Error(w, "ResponseWriter does not support CloseNotifier", http.StatusInternalServerError)
			return
		}
		cnc := cn.CloseNotify()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		ticker := time.NewTicker(d)
		for {
			select {
			case ev := <-ec:
				if ev.Event != "" {
					fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Event, ev.Data)
				} else {
					fmt.Fprintf(w, "data: %s\n\n", ev.Data)
				}
			case <-ticker.C:
				fmt.Fprintf(w, ": comment\n\n")
			case <-cnc:
				log.Println("closing connection from", req.RemoteAddr)
				fl.Flush()
				ticker.Stop()
				return
			}
			fl.Flush()
		}
	})
}

func imageHandler(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Query().Get(":key")

	var res *transformResult
	var found bool
	transforms.Lock()
	res, found = transforms.images[key]
	transforms.Unlock()

	if !found {
		http.Error(w, fmt.Sprintf("no image with key %s", key), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(res.Image)
}

var showTmplRaw = `<!DOCTYPE html>
<html>
<head><title>Primi Image</title></head>
<body>
  <img src="{{.ImgURL}}"/>
<div id="links">
  <a href="{{.SrcURL}}">Original Image</a><br>
  <a href="{{.ImgURL}}">Transformed Image</a>
</div>
</body>
</html>
`
var showTmpl = template.Must(template.New("primi").Parse(showTmplRaw))

func showHandler(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Query().Get(":key")

	var res *transformResult
	var found bool
	transforms.Lock()
	res, found = transforms.images[key]
	transforms.Unlock()

	if !found {
		http.Error(w, fmt.Sprintf("no image with key %s", key), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	showTmpl.Execute(w, res)
}

var addr = flag.String("a", ":8100", "server address")
var nworkers = flag.Int("n", 1, "transformation workers")
var commentd = flag.Duration("k", time.Duration(4)*time.Second, "keep alive sse duration")

func main() {
	log.SetPrefix("primi: ")
	flag.Parse()

	transforms.images = make(map[string]*transformResult)
	eventsCh := make(chan *event)
	// a buffered channel because we want the request handler to respond fast
	tc := make(chan *transformRequest, 100)

	for i := 0; i < *nworkers; i++ {
		go transformer(tc, eventsCh)
	}

	m := pat.New()
	m.Get("/primi", newSSEHandler(eventsCh, *commentd))
	m.Post("/images", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rawurl := req.FormValue("url")
		draw := req.FormValue("draw")
		if draw == "" {
			draw = "primitive"
		}
		tc <- &transformRequest{rawurl, draw}
		log.Printf("transformer accepted %q for %q", draw, rawurl)
		w.WriteHeader(http.StatusAccepted)
	}))
	m.Get("/image/:key", http.HandlerFunc(imageHandler))
	m.Get("/show/:key", http.HandlerFunc(showHandler))
	http.Handle("/", m)
	log.Println("Starting server")
	http.ListenAndServe(*addr, nil)
}

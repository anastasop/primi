package main

import (
	"image"
	"math/rand"
	"runtime"

	"github.com/fogleman/primitive/primitive"
	"github.com/nfnt/resize"
)

// primitiveOnImage is the main workhorse. Transforms an image
// with primitive and reasonable defaults
// TODO: the client should decide the options
func primitiveOnImage(img image.Image) image.Image {
	var (
		Count      = 300
		Alpha      = 128
		InputSize  = 256
		OutputSize = 800
		Mode       = rand.Intn(9)
		Workers    = runtime.NumCPU()
		Repeat     = 0
	)

	Count = (rand.Intn(17) + 4) * 50 // sth between 200 and 1000

	imgThumbnail := resize.Thumbnail(uint(InputSize), uint(InputSize), img, resize.Bilinear)
	backgroundColor := primitive.MakeColor(primitive.AverageImageColor(imgThumbnail))

	model := primitive.NewModel(imgThumbnail, backgroundColor, OutputSize, Workers)
	for i := 0; i < Count; i++ {
		model.Step(primitive.ShapeType(Mode), Alpha, Repeat)
	}

	return model.Context.Image()
}

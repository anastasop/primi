chrome.runtime.onInstalled.addListener(function () {
    chrome.contextMenus.create({
        "id": "primitive-draw",
        "title": "draw with primitive",
        "contexts": ['image']
    }, function () {
        if (chrome.extension.lastError) {
            console.log("contextMenu registration failed: ", chrome.extension.lastError.message);
        }
    });
    chrome.contextMenus.create({
        "id": "triangle-draw",
        "title": "draw with triangle",
        "contexts": ['image']
    }, function () {
        if (chrome.extension.lastError) {
            console.log("contextMenu registration failed: ", chrome.extension.lastError.message);
        }
    });
});

function getAddrPromise() {
    return new Promise(function (resolve, reject) {
        chrome.storage.local.get({
	    "addr": "http://localhost:8100"
	}, function (items) {
            if (chrome.extension.lastError) {
                reject(chrome.extension.lastError.message);
            } else {
                resolve(items["addr"]);
            }
        });
    })
}

chrome.runtime.onInstalled.addListener(function () {
    getAddrPromise().then(function (addr) {
        var evtSource = new EventSource(addr + "/primi");
        evtSource.addEventListener("image", function (e) {
            var msg = JSON.parse(e.data);
            chrome.notifications.create(null, {
                type: 'basic',
                iconUrl: 'primi.png',
                title: 'primi',
                message: msg.message,
                requireInteraction: true,
                isClickable: true
            }, function (notificationId) {
                chrome.notifications.onClicked.addListener(function (id) {
                    if (notificationId === id) {
                        chrome.notifications.clear(notificationId, function () {
                            chrome.tabs.create({ url: addr + msg.url, active: true })
                        })
                    }
                })
            })
        })
        evtSource.addEventListener("problem", function (e) {
            console.log(e.data)
        })
    }).catch(function (error) {
        console.log('request failed: ', error)
    });
})

chrome.contextMenus.onClicked.addListener(function (info) {
    var fetchOptions = {
        method: 'post',
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        }
    };

    if (info.menuItemId === "primitive-draw") {
        fetchOptions["body"] = "url=" + info.srcUrl + "&draw=primitive"
    } else if (info.menuItemId === "triangle-draw") {
        fetchOptions["body"] = "url=" + info.srcUrl + "&draw=triangle"
    }

    getAddrPromise().then(function (addr) {
        fetch(addr + "/images", fetchOptions).then(function (response) {
            if (response.status == 202) {
                return Promise.resolve(response)
            } else {
                return Promise.reject(new Error(response.statusText))
            }
        }).catch(function (error) {
            console.log('request failed: ', error)
        });
    });
});

chrome.browserAction.onClicked.addListener(function () {
    chrome.runtime.openOptionsPage(function () {
        if (chrome.extension.lastError) {
            console.log("Got expected error: " + chrome.extension.lastError.message)
        }
    });
});

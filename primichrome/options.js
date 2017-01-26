function restoreOptions() {
    chrome.storage.local.get({
        "addr": "http://localhost:8100"
    }, function (items) {
        if (!chrome.extension.lastError) {
            document.getElementById('addr').value = items["addr"];
        }
    });
}

function saveOptions() {
    var addr = document.getElementById('addr').value;
    chrome.storage.local.set({
        addr: addr.trim()
    }, function () {
        if (!chrome.extension.lastError) {
            restoreOptions();
        }
    });
}

document.addEventListener('DOMContentLoaded', restoreOptions);
document.getElementById('change').addEventListener('click', saveOptions);

Primi is a chrome extension for [primitive](https://primitive.lol/) a sophisticated image filter. Quoting the use from the site
> The user provides an image as input. The program tries to find the most optimal shape that can be drawn to maximize the similarity between the target image and the drawn image. It repeats this process, adding one shape at a time.

The results are very artistic. There is a [twitter bot](https://twitter.com/PrimitivePic) which you can follow for some great example. The project is [open source](https://github.com/fogleman/primitive) and hosted at github.

I like it very much but i don't have a mac to see the native app and using the command line app from github is very tedious. You see a nice image while browsing, download, save as, open terminal, run primitive, check console until it finished, open an image viewer for the result etc etc

To make it easier to use i wrote this chrome extension. It has a right click menu option to run primitive without leaving the browser. The extension also need a local server to run the transformations.

# Installation

1. `go get github.com/anastasop/primi/...` This install the server in `$GOPATH/bin/primiserver`
2. start the server in a terminal console. A script to start it automatically when booting is recommended
3. install the chrome extension from `$GOPATH/src/github.com/anastasop/primi/primichrome` If you are familiar with chrome, open the extensions page, set developer mode on and load it from the directory. More detailed instructions [here](https://developer.chrome.com/extensions/getstarted#unpacked). The server has an option for the primiserver. The default is `http://localhost:8100`. If you decide to change it, write the new address, click `Change` and reload the extension.

Now you are ready to use it. Browse the web and find a nice image like this ![this](./images/menu.png). Right click on it and click `draw with primitive`.


After a while a notification ![notification](./images/notification.png) will appear.


Click it and see the result of primitive ![primitive](./images/result.png)

# TODO

1. All images are stored in the server and expire after 5 minutes. It would be nice to use some kind of storage and an index page
2. Pack the server to run in google cloud engine and the extension for the chrome store

# Bugs

You must reload the extension if you change the server address in options.

# Icon

The icon of the chrome extension is the [Arles Cafe](https://en.wikipedia.org/wiki/Caf%C3%A9_Terrace_at_Night) as is today, filtered with primitive.
Enjoy!



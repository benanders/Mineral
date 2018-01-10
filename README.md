
Mineral
-------

Mineral is a Minecraft clone written in Go, using modern OpenGL. The goals of the project are:

* **Visually accurate**: the game attempts to be mostly true to the original in terms of the visual appearence, but allows for select changes and the addition of new features that may enhance usability (or make my life easier!)
* **Extensible**: the source code is well commented and documented, written in a clean, easily maintainable, extensible fashion
* **Modern**: apart from the fact that I'm using C, the implementation uses modern technologies and OpenGL rendering techniques
* **Technically unique**: the implementation is incompatible with the original game, differing in the design of its APIs, protocols, and architecture

This project is really just for my own amusement, so don't expect it to actually go anywhere!

## Building

You'll need SDL2 installed in order to build Mineral from scratch.
See the [README](https://github.com/veandco/go-sdl2) in the Go SDL2 bindings repository for a great description on how to install it.

Then just run:

```bash
$ go get -v github.com/benanders/Mineral
$ cd $GOPATH/src/github.com/benanders/Mineral
$ go run *.go
```

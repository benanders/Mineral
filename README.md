
Mineral
-------

Mineral is a Minecraft clone written in Go, using modern OpenGL. The goals of
the project are:

* **Visually accurate**: attempt to be mostly true to the original game in
  terms of visual appearence, but allows for select changes and the addition of
  new features that may enhance usability (or that make my life easier!)
* **Complete**: attempt to recreate the majority of Minecraft's most used
  features
* **Modern**: use modern technologies and OpenGL rendering techniques
* **Extensible**: well commented and documented source code, written in a
  clean, easily maintainable, extensible fashion
* **Technically unique**: the implementation is incompatible with the original
  game, differing in the design of its APIs, protocols, and architecture

This project is really just for my own amusement, so don't expect it to
actually go anywhere!

## Building

You'll need SDL2 installed in order to build Mineral from scratch.  See the
[README](https://github.com/veandco/go-sdl2) in the Go SDL2 bindings repository
for a great description on how to install it.

Then just run:

```bash
$ go get -v github.com/benanders/Mineral
$ cd $GOPATH/src/github.com/benanders/Mineral
$ go run *.go
```

## To Do

Core features:

[ ] Collision detection
[ ] Gravity, drag, other motion physics for the player
[ ] Texture rendering
[ ] More blocks
[ ] Wireframe block highlighting
[ ] Block placement, destruction
[ ] Calculate sky/block lighting values on separate goroutine
[ ] Smooth lighting
[ ] Ambient occlusion
[ ] Terrain generation
[ ] Jumping
[ ] Sprinting
[ ] Peaceful mobs (sheep, cows, etc.)
[ ] Violent mobs (zombies, creepers, etc.)
[ ] Special blocks (fences, stairs, half slabs, TNT, etc.)
[ ] Sounds (walking, mobs, etc.)
[ ] In game GUI
[ ] Player inventory
[ ] Player quick inventory
[ ] Tools
[ ] View bobbing
[ ] Crafting in player's inventory
[ ] Crafting table GUI
[ ] Crafting with crafting table

Additional features:

[ ] Main menu
[ ] Pause menu
[ ] In-game options
[ ] Game saving/loading

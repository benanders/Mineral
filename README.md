
Mineral
-------

Mineral is a Minecraft clone written in Go, using modern OpenGL. The goals of
the project are:

* **Visually accurate**: attempt to be mostly true to the original game in
  terms of visual appearence, but allows for select changes and the addition of
  new features that may enhance usability (or that make my life easier!).
* **Complete**: attempt to recreate the majority of Minecraft's most used
  features.
* **Modern**: use modern technologies and OpenGL rendering techniques.
* **Extensible**: well commented and documented source code, written in a
  clean, easily maintainable, extensible fashion.
* **Technically unique**: the implementation is incompatible with the original
  game, differing in the design of its APIs, protocols, and architecture.

This project is really just for my own amusement, so don't expect it to
actually go anywhere!

## Building

The [Mojang Brand and Asset
Guidelines](https://account.mojang.com/documents/brand_guidelines) are pretty
lenient for non-commercial things, but explicitly prohibit distribution of
their assets. So, you'll need to extract the assets from the original game
before being able to build Mineral. This means you'll need to own a copy of
Minecraft.

First, buy Minecraft from the [Mojang website](https://minecraft.net/en-us/),
install it, and run the game (not just the launcher) at least once to download
all the assets.

Next, clone this repository:

```bash
$ git clone https://github.com/benandrs/Mineral
$ cd Mineral
```

Then run the asset extraction script:

```bash
$ go run buildAssets.go
```

Install [go-bindata](https://github.com/jteeuwen/go-bindata):

```bash
$ go get -u github.com/jteeuwen/go-bindata/...
```

Compile all the assets into a single Go file for inclusion in the executable:

```bash
$ go-bindata -pkg asset -prefix "asset/data" -ignore "\.DS_Store" -o asset/asset.go asset/data/...
```

Install the Go SDL2 bindings using the instructions in the
[README](https://github.com/veandco/go-sdl2) of their repository.

Now you can run Mineral with:

```bash
$ go run main.go
```

## License

All the code that I've written here is under the MIT license, so you are pretty
much free to do whatever you want with it.

The [Mojang Brand and Asset
Guidelines](https://account.mojang.com/documents/brand_guidelines) require me
to say that Mineral is not an official Minecraft product, and is not approved
by or associated with Mojang in any way.

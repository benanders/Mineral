// Mineral is a Minecraft clone written in Go, using modern OpenGL. It aims to
// be visually accurate, extensible, modern, and technically unique.
package main

import (
	"log"
	"runtime"
	"time"

	"github.com/benanders/mineral/game"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

// The minimum number of nanoseconds that must elapse between update ticks.
const nsPerTick = 1000 * 1000 * 1000 / 60

func init() {
	// The OpenGL context MUST be created on the main OS thread. To ensure this,
	// we lock the main OS thread
	runtime.LockOSThread()
}

func main() {
	// Initialise SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalln("failed to initialise SDL:", err)
	}
	defer sdl.Quit()

	// Create a new window with a fixed title and initial size
	window, err := sdl.CreateWindow("Mineral",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 850, 500,
		sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_OPENGL|sdl.WINDOW_RESIZABLE)
	if err != nil {
		log.Fatalln("failed to create a new window:", err)
	}
	defer window.Destroy()

	// Trap the mouse cursor in the window
	sdl.SetRelativeMouseMode(true)

	// Hint the OpenGL version we want to use (v3.3 core)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	// Create the OpenGL context
	context, err := sdl.GLCreateContext(window)
	if err != nil {
		log.Fatalln("failed to create OpenGL context:", err)
	}
	defer sdl.GLDeleteContext(context)

	// Initialise OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialise OpenGL:", err)
	}

	// Print the OpenGL version in use
	glVersion := gl.GoStr(gl.GetString(gl.VERSION))
	glslVersion := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	log.Println("OpenGL version:", glVersion)
	log.Println("GLSL version:", glslVersion)

	// Create the main game state
	game := game.New(window)
	defer game.Destroy()

	// `lag` accumulates how much time each frame takes, so we can run the
	// update function at a constant time step
	previousTime := time.Now()
	lag := int64(0)

	// Main game loop
	running := true
	for running {
		// Calculate how long the last frame took
		currentTime := time.Now()
		elapsed := currentTime.Sub(previousTime).Nanoseconds()
		lag += elapsed
		previousTime = currentTime

		// Handle user input
		for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
			if _, ok := evt.(*sdl.QuitEvent); ok {
				running = false
			} else {
				game.HandleEvent(evt)
			}
		}

		// Update the game at a fixed time step, triggering multiple updates if
		// we've fallen behind (e.g. if rendering or the previous update takes
		// too long)
		for lag >= nsPerTick {
			game.Update()
			lag -= nsPerTick
		}

		// Render the game as fast as possible, dropping render frames to update
		// the game if necessary
		game.Render()
		sdl.GLSwapWindow(window)
	}
}

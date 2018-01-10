package entity

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

// InputCtrl controls an entity's movement and look direction based on
// user input from the keyboard and mouse.
type InputCtrl struct {
	walkSpeed      float32   // How fast the entity can walk
	lookSpeed      float32   // How fast the entity can look around
	IsKeyDown      [256]bool // Whether a key is pressed
	mouseX, mouseY int32     // Accumulates mouse movement over a frame
}

// NewInputCtrl creates a new input controller instance with the given walk
// and look speeds.
func NewInputCtrl(walkSpeed, lookSpeed float32) *InputCtrl {
	return &InputCtrl{walkSpeed: walkSpeed, lookSpeed: lookSpeed}
}

// HandleEvent updates its stored input state based off changes in user input.
func (c *InputCtrl) HandleEvent(evt sdl.Event) {
	switch e := evt.(type) {
	case *sdl.KeyboardEvent:
		if e.Keysym.Scancode < 256 {
			c.IsKeyDown[e.Keysym.Scancode] = (e.State == sdl.PRESSED)
		}
	case *sdl.MouseMotionEvent:
		c.mouseX += e.XRel
		c.mouseY += e.YRel
	}
}

// Update modifies an entity's position and look direction based on the keyboard
// and mouse input that has been detected over the past frame.
func (c *InputCtrl) Update(entity *Entity) {
	// Update position based on keyboard input
	x, y, z := float32(0.0), float32(0.0), float32(0.0)
	if c.IsKeyDown[sdl.SCANCODE_W] {
		z += c.walkSpeed
	}
	if c.IsKeyDown[sdl.SCANCODE_S] {
		z -= c.walkSpeed
	}
	if c.IsKeyDown[sdl.SCANCODE_A] {
		x -= c.walkSpeed
	}
	if c.IsKeyDown[sdl.SCANCODE_D] {
		x += c.walkSpeed
	}
	if c.IsKeyDown[sdl.SCANCODE_SPACE] {
		y += c.walkSpeed
	}
	if c.IsKeyDown[sdl.SCANCODE_LSHIFT] || c.IsKeyDown[sdl.SCANCODE_RSHIFT] {
		y -= c.walkSpeed
	}
	entity.Walk(mgl32.Vec3{x, y, z})

	// Update look direction based on mouse input
	verticalDelta := float32(c.mouseY) * c.lookSpeed
	horizontalDelta := float32(c.mouseX) * c.lookSpeed
	entity.Look(mgl32.Vec2{horizontalDelta, verticalDelta})
	c.mouseX, c.mouseY = 0.0, 0.0
}

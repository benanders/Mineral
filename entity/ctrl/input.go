package ctrl

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

// InputCtrl controls an entity's movement and look direction based on user
// input from the keyboard and mouse.
type InputCtrl struct {
	IsKeyDown      [256]bool // Whether a key is pressed
	mouseX, mouseY int32     // Accumulates mouse movement over a frame
}

// NewInputCtrl creates a new input controller instance with the given walk and
// look speeds.
func NewInputCtrl() *InputCtrl {
	return &InputCtrl{}
}

// HandleEvent implements the `Controller` interface.
func (c *InputCtrl) HandleEvent(evt sdl.Event) {
	switch e := evt.(type) {
	case *sdl.KeyboardEvent:
		// Prevent an index out of bounds error
		if e.Keysym.Scancode < 256 {
			c.IsKeyDown[e.Keysym.Scancode] = (e.State == sdl.PRESSED)
		}
	case *sdl.MouseMotionEvent:
		c.mouseX += e.XRel
		c.mouseY += e.YRel
	}
}

// Update implements the `Controller` interface.
func (c *InputCtrl) Update(entity Controllable) {
	// Update position based on keyboard input
	x, y, z := float32(0.0), float32(0.0), float32(0.0)
	if c.IsKeyDown[sdl.SCANCODE_W] {
		z += 1.0
	}
	if c.IsKeyDown[sdl.SCANCODE_S] {
		z -= 1.0
	}
	if c.IsKeyDown[sdl.SCANCODE_A] {
		x -= 1.0
	}
	if c.IsKeyDown[sdl.SCANCODE_D] {
		x += 1.0
	}
	if c.IsKeyDown[sdl.SCANCODE_SPACE] {
		y += 1.0
	}
	if c.IsKeyDown[sdl.SCANCODE_LSHIFT] || c.IsKeyDown[sdl.SCANCODE_RSHIFT] {
		y -= 1.0
	}
	entity.Walk(mgl32.Vec3{x, y, z})

	// Update the entity's look direction based on mouse input
	horizontalDelta := float32(c.mouseX)
	verticalDelta := float32(c.mouseY)
	entity.Look(mgl32.Vec2{horizontalDelta, verticalDelta})
	c.mouseX, c.mouseY = 0.0, 0.0
}

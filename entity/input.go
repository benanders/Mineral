package entity

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

// Controllable is implemented by all entities that can be controlled with a
// controller (e.g. the input controller, or one of the AI controllers).
type Controllable interface {
	// Move moves the entity an amount forwards, right, and up. Delta values
	// are normalized to 1, and should be multiplied by the entity's move speed
	// prior to applying the movement.
	Move(delta mgl32.Vec3)

	// Look modifies the look direction of an entity by an amount. Delta values
	// are normalized to 1, and should be multiplied by the entity's look speed
	// prior to applying the rotation.
	Look(delta mgl32.Vec2)
}

// Controller is implemented by all entity controllers (e.g. the input
// controller, or the mob AI controllers).
type Controller interface {
	// HandleEvent is called whenever a user event is triggered (used by the
	// input controller).
	HandleEvent(evt sdl.Event)

	// Update is called every frame to modify an entity's position and look
	// direction.
	Update(entity Controllable)
}

// InputCtrl controls an entity's movement and look direction based on user
// input from the keyboard and mouse.
type InputCtrl struct {
	IsKeyDown      [256]bool // Whether a key is pressed
	mouseX, mouseY int32     // Accumulates mouse movement over a frame
}

// NewInputCtrl creates a new input controller instance with the given move and
// look speeds.
func NewInputCtrl() *InputCtrl {
	return &InputCtrl{}
}

// HandleEvent implements the `Controller` interface.
func (c *InputCtrl) HandleEvent(evt sdl.Event) {
	switch e := evt.(type) {
	case *sdl.KeyboardEvent:
		// Prevent an index out of bounds error
		if int(e.Keysym.Scancode) < len(c.IsKeyDown) {
			c.IsKeyDown[e.Keysym.Scancode] = (e.State == sdl.PRESSED)
		}
	case *sdl.MouseMotionEvent:
		c.mouseX += e.XRel
		c.mouseY += e.YRel
	}
}

// Update implements the `Controller` interface.
func (c *InputCtrl) Update(entity Controllable) {
	// Update the entity's look direction based on mouse input. We do this
	// first so that the entity's local coordinate system is updated before
	// applying movement
	horizontalDelta := float32(c.mouseX)
	verticalDelta := float32(c.mouseY)
	entity.Look(mgl32.Vec2{horizontalDelta, verticalDelta})
	c.mouseX, c.mouseY = 0.0, 0.0

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
	entity.Move(mgl32.Vec3{x, y, z})
}

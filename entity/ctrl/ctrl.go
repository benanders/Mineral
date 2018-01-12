package ctrl

// The `ctrl` package (shortened version of `control`) provides mechanisms for
// controlling entity movement and look direction. It contains all mob AI
// controllers and the player's input controller.

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

// Controllable is implemented by all entities that can be controlled with a
// controller (e.g. the input controller, or one of the AI controllers).
type Controllable interface {
	// Move moves the entity an amount forwards, right, and up. Delta values
	// are normalised to 1, and should be multiplied by the entity's move speed
	// prior to applying the movement.
	Move(delta mgl32.Vec3)

	// Look modifies the look direction of an entity by an amount. Delta values
	// are normalised to 1, and should be multiplied by the entity's look speed
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

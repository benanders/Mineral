package entity

import (
	"math"

	"github.com/chewxy/math32"
	"github.com/go-gl/mathgl/mgl32"
)

// Entity represents a movable creature in the game world (e.g. the player,
// NPCs, animals, mobs, etc).
type Entity struct {
	position mgl32.Vec3 // Position of the entity in world coordinates
	rotation mgl32.Vec2 // Viewing direction along the x and y axes
	sight    mgl32.Vec3 // Points in the direction the entity is looking
	forward  mgl32.Vec3 // Points in the direction the entity walks
	right    mgl32.Vec3 // Points in the direction the entity strafes
	up       mgl32.Vec3 // Points in the direction the entity can fly
}

// NewEntity creates a new instance of the entity with an initial position and
// look direction.
func NewEntity(position mgl32.Vec3, rotation mgl32.Vec2) *Entity {
	e := Entity{}
	e.position = position
	e.rotation = rotation
	e.updateAxes()
	return &e
}

// Sight implements the Entity interface for the entity. It returns a vector
// that points in the direction the entity is looking.
func (e *Entity) Sight() mgl32.Vec3 {
	return e.sight
}

// Position implements the Entity interface for the entity. It returns the
// entity's position in world coordinates.
func (e *Entity) Position() mgl32.Vec3 {
	return e.position
}

// Walk moves the entity forward, right, and up by a certain amount in its
// local coordinate basis.
func (e *Entity) Walk(delta mgl32.Vec3) {
	// forward, right, and up form an orthonormal basis in the entity's
	// coordinate system
	forward := e.forward.Mul(delta.Z())
	right := e.right.Mul(delta.X())
	up := e.up.Mul(delta.Y())
	e.position = e.position.Add(forward.Add(right.Add(up)))
}

// Look rotates the entity's look direction by a certain amount in the
// horizontal and vertical directions.
func (e *Entity) Look(delta mgl32.Vec2) {
	// Clamp the vertical look direction
	y := e.rotation.Y() + delta.Y()
	if y >= math.Pi/2.0-0.0001 {
		y = math.Pi/2.0 - 0.0001
	} else if y <= -math.Pi/2.0+0.0001 {
		y = -math.Pi/2.0 + 0.0001
	}

	// Update the entity's look direction
	e.rotation = mgl32.Vec2{e.rotation.X() + delta.X(), y}
	e.updateAxes()
}

// UpdateAxes recalculates the entity's orthonormal basis formed by forward,
// right, and up, based on the entity's current look direction.
func (e *Entity) updateAxes() {
	sinX, cosX := math32.Sincos(e.rotation.X())
	sinY, cosY := math32.Sincos(e.rotation.Y())
	e.forward = mgl32.Vec3{sinX, 0.0, -cosX}
	e.right = mgl32.Vec3{cosX, 0.0, sinX}
	e.up = mgl32.Vec3{0.0, 1.0, 0.0}
	e.sight = mgl32.Vec3{cosY * -sinX, sinY, cosY * cosX}
}

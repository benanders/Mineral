package entity

import (
	"math"

	"github.com/benanders/mineral/util"

	"github.com/chewxy/math32"
	"github.com/go-gl/mathgl/mgl32"
)

// Entity represents a movable creature in the game world (e.g. the player,
// NPCs, animals, mobs, etc).
//
// An entity's position and size is specified by an axis aligned bounding box
// (AABB), and its rotation by a 2D vector. The AABB is used for collision
// detection, and is not affected by the entity's rotation. The rotation only
// affects rendering, and specifies the entity's viewing direction (i.e. sight
// vector) in spherical coordinates.
type Entity struct {
	AABB     util.AABB  // AABB specifying position and size
	Rotation mgl32.Vec2 // Rotation along the x and y axes

	Sight   mgl32.Vec3 // Points in the direction the entity is looking
	forward mgl32.Vec3 // Points in the direction the entity walks
	right   mgl32.Vec3 // Points in the direction the entity strafes
	up      mgl32.Vec3 // Points in the direction the entity can fly

	walkSpeed float32 // The speed at which the entity can move around
	lookSpeed float32 // The speed at which the entity can look around
}

// NewEntity creates a new instance of the entity with an initial position,
// size (specified by the entity's AABB), and rotation.
func NewEntity(aabb util.AABB, rotation mgl32.Vec2, walkSpeed,
	lookSpeed float32) *Entity {
	e := Entity{AABB: aabb, Rotation: rotation, walkSpeed: walkSpeed,
		lookSpeed: lookSpeed}
	e.updateAxes()
	return &e
}

// Walk moves the entity forward, right, and up by a certain amount in its
// local coordinate basis.
//
// Implements the `ctrl.Controllable` interface.
func (e *Entity) Walk(delta mgl32.Vec3) {
	// Calculate how much we need to move along each of the entity's axes based
	// on the delta
	forward := e.forward.Mul(delta.Z() * e.walkSpeed)
	right := e.right.Mul(delta.X() * e.walkSpeed)
	up := e.up.Mul(delta.Y() * e.walkSpeed)

	// Update the entity's position by summing the movements along each axis
	e.AABB.Offset(forward.Add(right.Add(up)))
}

// Look rotates the entity's look direction by a certain amount in the
// horizontal and vertical directions.
//
// Implements the `ctrl.Controllable` interface.
func (e *Entity) Look(delta mgl32.Vec2) {
	x := e.Rotation.X() + delta.X()*e.lookSpeed
	y := e.Rotation.Y() + delta.Y()*e.lookSpeed

	// Clamp the vertical look direction; use a small epsilon since the
	// rendering seems to screw up if we don't
	epsilon := float32(0.0001)
	y = util.Clamp(y, -math.Pi/2.0+epsilon, math.Pi/2.0-epsilon)

	// Update the entity's rotation and orthonormal movement axes based on the
	// new look direction
	e.Rotation = mgl32.Vec2{x, y}
	e.updateAxes()
}

// UpdateAxes recalculates the entity's orthonormal basis formed by forward,
// right, and up, based on the entity's current look direction.
func (e *Entity) updateAxes() {
	// The movement vectors are calculated as a conversion from cylindrical to
	// rectangular Cartesian coordinates
	sinX, cosX := math32.Sincos(e.Rotation.X())
	e.forward = mgl32.Vec3{sinX, 0.0, -cosX}
	e.right = mgl32.Vec3{cosX, 0.0, sinX}
	e.up = mgl32.Vec3{0.0, 1.0, 0.0}

	// The sight vector is calculated as a conversion from spherical to
	// rectangular Cartesian coordinates
	sinY, cosY := math32.Sincos(e.Rotation.Y())
	e.Sight = mgl32.Vec3{cosY * -sinX, sinY, cosY * cosX}
}

package entity

import (
	"github.com/benanders/mineral/math"
	"github.com/benanders/mineral/world"

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
	AABB     math.AABB  // AABB specifying position and size
	Rotation mgl32.Vec2 // Rotation along the x and y axes

	Sight   mgl32.Vec3 // Points in the direction the entity is looking
	forward mgl32.Vec3 // Points in the direction the entity moves
	right   mgl32.Vec3 // Points in the direction the entity strafes
	up      mgl32.Vec3 // Points in the direction the entity can fly

	moveSpeed float32 // The speed at which the entity can move around
	lookSpeed float32 // The speed at which the entity can look around

	// We aggregate all movement over an update tick before applying the
	// movement delta and performing collision detection.
	//
	// `moveDelta` is in world coordinate space, so we can just sum the current
	// position with the delta to get the new position.
	moveDelta mgl32.Vec3
}

// NewEntity creates a new instance of the entity with an initial position,
// size (specified by the entity's AABB), and rotation.
func NewEntity(aabb math.AABB, rotation mgl32.Vec2, moveSpeed,
	lookSpeed float32) *Entity {
	e := Entity{AABB: aabb, Rotation: rotation, moveSpeed: moveSpeed,
		lookSpeed: lookSpeed}
	e.updateAxes()
	return &e
}

// Move moves the entity forward, right, and up by a certain amount in its
// local coordinate basis.
//
// Implements the `ctrl.Controllable` interface.
func (e *Entity) Move(delta mgl32.Vec3) {
	// Calculate how much we need to move along each of the entity's axes based
	// on the delta
	forward := e.forward.Mul(delta.Z() * e.moveSpeed)
	right := e.right.Mul(delta.X() * e.moveSpeed)
	up := e.up.Mul(delta.Y() * e.moveSpeed)

	// Calculate the delta in world coordinates by summing the deltas along the
	// 3 entity axes
	worldDelta := forward.Add(right.Add(up))
	e.moveDelta = e.moveDelta.Add(worldDelta)
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
	y = math.Clamp(y, -math32.Pi/2.0+epsilon, math32.Pi/2.0-epsilon)

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

// CollisionAxis represents an axis along which we can resolve a collision.
type collisionAxis uint

const (
	// The three possible collision axes are the x, y, and z axes.
	axisX collisionAxis = iota
	axisY
	axisZ
)

// ApplyMovementAndResolveCollisions applies the accumulated movement delta
// that's been collected since the previous update tick, and resolves
// collisions between the entity and all solid blocks in the world.
func (e *Entity) ApplyMovementAndResolveCollisions(w *world.World) {
	// X axis
	e.AABB.Offset(mgl32.Vec3{e.moveDelta.X(), 0.0, 0.0})
	e.resolveBlockCollisions(w, axisX)

	// Y axis
	e.AABB.Offset(mgl32.Vec3{0.0, e.moveDelta.Y(), 0.0})
	e.resolveBlockCollisions(w, axisY)

	// Z axis
	e.AABB.Offset(mgl32.Vec3{0.0, 0.0, e.moveDelta.Z()})
	e.resolveBlockCollisions(w, axisZ)

	// Reset the movement delta
	e.moveDelta = mgl32.Vec3{}
}

// ResolveBlockCollisions checks to see if the entity is colliding with any
// solid blocks in the world, and if so resolves the collision by moving
// the entity along the specified axis.
func (e *Entity) resolveBlockCollisions(w *world.World, axis collisionAxis) {
	// Calculate the bounds of the entity's AABB in block coordinates
	x1, y1, z1 := world.ToWorldSpace(e.AABB.MinX(), e.AABB.MinY(),
		e.AABB.MinZ())
	x2, y2, z2 := world.ToWorldSpace(e.AABB.MaxX(), e.AABB.MaxY(),
		e.AABB.MaxZ())

	// Iterate over all blocks that overlap the entity
	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			for z := z1; z <= z2; z++ {
				e.resolveBlockCollision(w, axis, x, y, z)
			}
		}
	}
}

// ResolveBlockCollision checks to see if the entity is colliding with the
// given block, and if so resolves the collision with this block by moving
// the entity along the specified axis.
func (e *Entity) resolveBlockCollision(w *world.World, axis collisionAxis,
	x, y, z int) {
	// Get the chunk containing the block
	p, q, cx, cy, cz := world.ToChunkSpace(x, y, z)
	chunk := w.FindChunk(p, q)

	// Don't bother detecting collisions with chunks that haven't loaded
	if chunk == nil || chunk.Blocks == nil {
		return
	}

	// Get the block we're checking for collisions against
	block := chunk.Blocks.At(cx, cy, cz)
	if block == nil {
		return
	}

	// Check the block we're colliding against is solid
	info := w.GetBlockInfo(*block)
	if !info.Collidable {
		return
	}

	// Resolve a collision with the block
	aabb := info.AABB(p, q, cx, cy, cz)
	e.resolveCollision(aabb, axis)
}

// ResolveCollision checks to see if the entity is colliding with the given
// AABB, and if so resolves the collision by moving the entity along the
// specified axis.
func (e *Entity) resolveCollision(other math.AABB, axis collisionAxis) {
	// Check the entity's AABB intersects the other AABB
	if !e.AABB.Intersects(other) {
		return
	}

	// Resolve the collision along the specified axis
	var offset mgl32.Vec3
	if axis == axisX {
		offset = mgl32.Vec3{-e.AABB.IntersectionX(other), 0.0, 0.0}
	} else if axis == axisY {
		offset = mgl32.Vec3{0.0, -e.AABB.IntersectionY(other), 0.0}
	} else if axis == axisZ {
		offset = mgl32.Vec3{0.0, 0.0, -e.AABB.IntersectionZ(other)}
	}
	e.AABB.Offset(offset)
}

package entity

import (
	"math"

	"github.com/benanders/mineral/util"
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
	AABB     util.AABB  // AABB specifying position and size
	Rotation mgl32.Vec2 // Rotation along the x and y axes

	Sight   mgl32.Vec3 // Points in the direction the entity is looking
	forward mgl32.Vec3 // Points in the direction the entity moves
	right   mgl32.Vec3 // Points in the direction the entity strafes
	up      mgl32.Vec3 // Points in the direction the entity can fly

	moveSpeed float32 // The speed at which the entity can move around
	lookSpeed float32 // The speed at which the entity can look around

	// Since we apply entity movement along each axis separately in order to
	// perform collision detection, we aggregate all movement over an update
	// tick before applying the movement delta and performing collision
	// detection.
	//
	// `moveDelta` is in world coordinate space, so we can just sum the current
	// position with the delta to get the new position.
	moveDelta mgl32.Vec3
}

// NewEntity creates a new instance of the entity with an initial position,
// size (specified by the entity's AABB), and rotation.
func NewEntity(aabb util.AABB, rotation mgl32.Vec2, moveSpeed,
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

// ApplyMovementAndResolveCollisions is called once per update tick to apply
// the movement delta that has accumulated since the last update frame, and
// resolve collisions with the world and other entities.
//
// See https://gamedev.stackexchange.com/a/71123 for an explanation on how
// collision detection is performed.
func (e *Entity) ApplyMovementAndResolveCollisions(world *world.World) {
	// Only apply movement if there was some movement
	epsilon := float32(0.00001)
	if math32.Abs(e.moveDelta.X()) < epsilon &&
		math32.Abs(e.moveDelta.Y()) < epsilon &&
		math32.Abs(e.moveDelta.Z()) < epsilon {
		return
	}

	// A list of the 3 axes we have to apply movement to and resolve collisions
	// along
	axes := [...]mgl32.Vec3{
		mgl32.Vec3{1.0, 0.0, 0.0},
		mgl32.Vec3{0.0, 1.0, 0.0},
		mgl32.Vec3{0.0, 0.0, 1.0}}

	// Iterate over each axis
	for _, axis := range axes {
		// Move the player along the axis
		e.AABB.Offset(mgl32.Vec3{
			e.moveDelta.X() * axis.X(),
			e.moveDelta.Y() * axis.Y(),
			e.moveDelta.Z() * axis.Z()})

		// Resolve collisions along the axis
		e.resolveBlockCollisions(world, axis)
	}

	// Reset the movement delta now that it's been applied
	e.moveDelta = mgl32.Vec3{}
}

// ResolveBlockCollisions checks to see if the entity is colliding with any
// blocks in the world, and moves the entity out of those blocks along the
// given axis.
func (e *Entity) resolveBlockCollisions(world *world.World, axis mgl32.Vec3) {
	// Calculate the bounds of the entity's AABB in block coordinates
	ax, bx := math32.Floor(e.AABB.MinX()), math32.Floor(e.AABB.MaxX())
	ay, by := math32.Floor(e.AABB.MinY()), math32.Floor(e.AABB.MaxY())
	az, bz := math32.Floor(e.AABB.MinZ()), math32.Floor(e.AABB.MaxZ())

	// Iterate over all blocks that overlap the entity
	for x := int(ax); x <= int(bx); x++ {
		for y := int(ay); y <= int(by); y++ {
			for z := int(az); z <= int(bz); z++ {
				e.resolveBlockCollision(world, axis, x, y, z)
			}
		}
	}
}

// ResolveBlockCollision checks to see if the entity is colliding with the
// block at the given coordinate, and if so moves the entity out of the block.
func (e *Entity) resolveBlockCollision(w *world.World, axis mgl32.Vec3,
	x, y, z int) {
	// Get the chunk containing the block
	p, q, cx, cy, cz := world.Chunked(x, y, z)
	chunk := w.FindChunk(p, q)

	// Don't bother detecting collisions if the chunk's block data isn't loaded
	if chunk == nil || chunk.Blocks == nil {
		return
	}

	// Check the block we're attempting to collide with is actually collidable
	block := chunk.Blocks.At(cx, cy, cz)
	if !block.IsCollidable() {
		return
	}

	// Resolve a collision with the block
	blockAABB := block.AABB(p, q, cx, cy, cz)
	e.resolveCollision(blockAABB, axis)
}

// ResolveCollision checks to see if the entity is colliding with the given
// AABB, and if so moves the entity out of that AABB.
func (e *Entity) resolveCollision(other util.AABB, axis mgl32.Vec3) {
	// Don't do anything if the AABBs don't overlap
	if !e.AABB.IsOverlapping(other) {
		return
	}

	// Resolve along the desired axis
	// println("axis", axis.X(), axis.Y(), axis.Z())
	depth := e.AABB.Overlap(other)
	// println("depth", depth.X(), depth.Y(), depth.Z())
	resolution := mgl32.Vec3{
		-depth.X() * axis.X(),
		-depth.Y() * axis.Y(),
		-depth.Z() * axis.Z()}
	println("resolution", resolution.X(), resolution.Y(), resolution.Z())
	e.AABB.Offset(resolution)
}

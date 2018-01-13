package world

import (
	"github.com/benanders/mineral/util"
	"github.com/go-gl/mathgl/mgl32"
)

// BlockData represents an array of blocks within a chunk.
type BlockData []BlockType

// NewBlockData creates a new blocks array for a chunk, with length equal to
// the number of blocks within a chunk.
func newBlockData() BlockData {
	return make([]BlockType, ChunkWidth*ChunkHeight*ChunkDepth)
}

// At returns the block at the given coordinate within the block list. If the
// given coordinates are outside the block list's boundaries, then returns
func (b BlockData) At(x, y, z int) *BlockType {
	if x < 0 || x >= ChunkWidth || y < 0 || y >= ChunkHeight || z < 0 ||
		z >= ChunkDepth {
		// Return an air block if the coordinate is outside the block data's
		// available range
		temp := BlockAir
		return &temp
	} else {
		return &b[y*ChunkWidth*ChunkDepth+z*ChunkWidth+x]
	}
}

// BlockType is the type of a block within the world.
type BlockType uint

// All block types.
const (
	BlockAir BlockType = iota
	BlockBedrock
	BlockStone
	BlockDirt
	BlockGrass
)

// IsTransparent tells us whether a block is at least partially transparent or
// not.
func (b BlockType) IsTransparent() bool {
	lookup := [...]bool{
		true,  // Air
		false, // Stone
		false, // Dirt
		false, // Grass
	}
	return lookup[b]
}

// IsCollidable tells us whether we have to test for collisions against a
// block.
func (b BlockType) IsCollidable() bool {
	lookup := [...]bool{
		false, // Air
		true,  // Stone
		true,  // Dirt
		true,  // Grass
	}
	return lookup[b]
}

// AABB returns an axis aligned bounding box for the block that we can check
// for collisions against.
//
// Blocks return different AABBs depending on their type (e.g. fences). This
// function is only ever called for blocks that are collidable.
func (b BlockType) AABB(p, q, x, y, z int) util.AABB {
	// We haven't implemented any special blocks yet, so just return a cube
	// at the given location
	rx, ry, rz := float32(p*ChunkWidth+x)+0.5, float32(y)+0.5,
		float32(q*ChunkDepth+z)+0.5
	return util.AABB{
		Center: mgl32.Vec3{rx, ry, rz},
		Size:   mgl32.Vec3{1.0, 1.0, 1.0}}
}

// BlockFace represents one of the possible 6 faces of a block.
type BlockFace uint

// All block faces.
const (
	FaceLeft BlockFace = iota
	FaceRight
	FaceTop
	FaceBottom
	FaceFront
	FaceBack
)

// Normal tells us the normal vector for a face.
func (f BlockFace) normal() (int, int, int) {
	lookup := [...][3]int{
		{-1, 0, 0}, // Left
		{1, 0, 0},  // Right
		{0, 1, 0},  // Top
		{0, -1, 0}, // Bottom
		{0, 0, 1},  // Front
		{0, 0, -1}, // Back
	}
	return lookup[f][0], lookup[f][1], lookup[f][2]
}

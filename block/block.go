package block

import (
	"github.com/benanders/mineral/math"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	ChunkWidth  = 16  // How wide a chunk is, in blocks.
	ChunkHeight = 256 // How high a chunk is, in blocks.
	ChunkDepth  = 16  // How deep a chunk is, in blocks.
)

// Block represents a block within the world, represented by a simple block ID
// assigned at game load.
type Block uint32

// Visible tells us whether a block actually draws something to the screen
// when present. So far, the only invisible block is air.
func (b Block) Visible() bool {
	return blockVariants[b].Visible
}

// Transparent tells us whether a block is at least partially transparent or
// not (whether we can see the block behind through this block).
func (b Block) Transparent() bool {
	return blockVariants[b].Transparent
}

// Collidable tells us whether we have to test for collisions against a block.
func (b Block) Collidable() bool {
	return blockVariants[b].Collidable
}

// AABB returns an axis aligned bounding box for the block, used for collision
// detection.
func (b Block) AABB(p, q, x, y, z int) math.AABB {
	// Add 0.5 since the AABB struct requires we specify the centre of the
	// block, and blocks are always 1x1 units
	rx := float32(p*ChunkWidth+x) + 0.5
	ry := float32(y) + 0.5
	rz := float32(q*ChunkDepth+z) + 0.5
	return math.AABB{
		Center: mgl32.Vec3{rx, ry, rz},
		Size:   mgl32.Vec3{1.0, 1.0, 1.0},
	}
}

// UV returns the coordinate in the texture atlas of the 16x16 pixel block to
// use to render a particular face of the block.
func (b Block) UV() BlockUV {
	return blockUVs[b]
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

// FaceNormals is an array indexed by block face that tells us the normal
// vector for each face.
var faceNormals = [...][3]int{
	{-1, 0, 0}, // Left
	{1, 0, 0},  // Right
	{0, 1, 0},  // Top
	{0, -1, 0}, // Bottom
	{0, 0, 1},  // Front
	{0, 0, -1}, // Back
}

// Normal tells us the normal vector for a face.
func (f BlockFace) Normal() (int, int, int) {
	return faceNormals[f][0], faceNormals[f][1], faceNormals[f][2]
}

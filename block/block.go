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

// Block represents the ID of a block within the world.
type Block uint32

// All available block variants.
const (
	Air Block = iota
	Bedrock
	Stone
	Dirt
	Grass
)

// IsTransparent tells us whether a block is at least partially transparent or
// not.
func (b Block) IsTransparent() bool {
	return BlockVariants[b].transparent
}

// IsCollidable tells us whether we have to test for collisions against a
// block.
func (b Block) IsCollidable() bool {
	return BlockVariants[b].collidable
}

// AABB returns an axis aligned bounding box for the block, used for collision
// detection.
func (b Block) AABB(p, q, x, y, z int) math.AABB {
	rx, ry, rz := float32(p*ChunkWidth+x)+0.5, float32(y)+0.5,
		float32(q*ChunkDepth+z)+0.5
	return math.AABB{
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
func (f BlockFace) Normal() (int, int, int) {
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

// BlockVariant contains the properties of a block type.
type blockVariant struct {
	name        string
	visible     bool // True if the block actually renders something
	collidable  bool // True if the block has a collidable AABB
	transparent bool // True if we can see the block behind this one

	// String containing path to image to use for all faces, or array of 6
	// strings containing paths to the images to use for each face, or nil if
	// the block is invisible.
	faces interface{}
}

// Lists all available block variants, and the information associated with
// them.
//
// The faces are listed in order of left, right, top, bottom, front, back.
var BlockVariants = [...]blockVariant{
	{
		name:        "Air",
		visible:     false,
		collidable:  false,
		transparent: true,
		faces:       nil,
	},
	{
		name:        "Bedrock",
		visible:     true,
		collidable:  true,
		transparent: false,
		faces:       "assets/minecraft/textures/blocks/bedrock.png",
	},
	{
		name:        "Stone",
		visible:     true,
		collidable:  true,
		transparent: false,
		faces:       "assets/minecraft/textures/blocks/stone.png",
	},
	{
		name:        "Dirt",
		visible:     true,
		collidable:  true,
		transparent: false,
		faces:       "assets/minecraft/textures/blocks/dirt.png",
	},
	{
		name:        "Grass",
		visible:     true,
		collidable:  true,
		transparent: false,
		faces: [...]string{
			"assets/minecraft/textures/blocks/grass_side.png",
			"assets/minecraft/textures/blocks/grass_side.png",
			"assets/minecraft/textures/blocks/grass_top.png",
			"assets/minecraft/textures/blocks/grass_side.png",
			"assets/minecraft/textures/blocks/grass_side.png",
			"assets/minecraft/textures/blocks/grass_side.png",
		},
	},
}

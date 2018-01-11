package world

// BlockData represents an array of blocks within a chunk.
type blockData []blockType

// NewBlockData creates a new blocks array for a chunk, with length equal to
// the number of blocks within a chunk.
func newBlockData() blockData {
	return make([]blockType, ChunkWidth*ChunkHeight*ChunkDepth)
}

// Creates a new blocks list filled with air,
// At returns the block at the given coordinate within the block list. If the
// given coordinates are outside the block list's boundaries, then returns
func (b blockData) at(x, y, z int) *blockType {
	if x < 0 || x >= ChunkWidth || y < 0 || y >= ChunkHeight || z < 0 ||
		z >= ChunkDepth {
		temp := blockAir
		return &temp
	} else {
		return &b[y*ChunkWidth*ChunkDepth+z*ChunkWidth+x]
	}
}

// BlockType is the type of a block within the world.
type blockType uint

// All block types.
const (
	blockAir blockType = iota
	blockStone
	blockDirt
	blockGrass
)

// IsTransparent tells us whether a block is at least partially transparent or
// not.
func (b blockType) isTransparent() bool {
	lookup := [...]bool{
		true,  // Air
		false, // Stone
		false, // Dirt
		false, // Grass
	}
	return lookup[b]
}

// BlockFace represents one of the possible 6 faces of a block, numbered from 0
// to 5.
type blockFace uint

// All block faces.
const (
	faceLeft blockFace = iota
	faceRight
	faceTop
	faceBottom
	faceFront
	faceBack
)

// Normal tells us the normal vector for a face.
func (f blockFace) normal() (int, int, int) {
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

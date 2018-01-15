package world

import "github.com/benanders/mineral/block"

// GenBlocks takes the coordinates for a chunk and procedurally generates the
// chunk's block data.
func genBlocks(p, q int) BlockData {
	// Create the block array
	blocks := newBlockData()

	// Populate the bottom 3 layers with stone
	for x := 0; x < block.ChunkWidth; x++ {
		for y := 0; y < 3; y++ {
			for z := 0; z < block.ChunkDepth; z++ {
				*blocks.At(x, y, z) = block.Stone
			}
		}
	}

	return blocks
}

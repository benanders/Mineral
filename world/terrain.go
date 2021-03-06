package world

// BlockGenInfo contains the necessary information to generate the terrain data
// for a chunk.
type blockGenInfo struct {
	p, q int // The location of the chunk to generate terrain data for
}

// GenBlocks takes the coordinates for a chunk and procedurally generates the
// chunk's block data.
func genBlocks(p, q int) blockData {
	// Create the block array
	blocks := newBlockData()

	// Populate the bottom 3 layers with stone
	for x := 0; x < ChunkWidth; x++ {
		for y := 0; y < 3; y++ {
			for z := 0; z < ChunkDepth; z++ {
				*blocks.At(x, y, z) = 3
			}
		}
	}

	return blocks
}

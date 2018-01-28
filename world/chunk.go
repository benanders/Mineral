package world

import (
	"github.com/go-gl/gl/v3.3-core/gl"
)

// The size of the chunk (width, height, and depth), in blocks.
const (
	ChunkWidth  = 16
	ChunkHeight = 256
	ChunkDepth  = 16
)

// ChunkPos represents the position of a chunk as a pair of x, z values
// (labelled p, q to distinguish between chunk and block coordinates).
type chunkPos struct {
	p, q int
}

// Chunk stores information associated with a chunk, including OpenGL rendering
// information, block data, vertex data, and lighting data.
type Chunk struct {
	P, Q        int       // The position of the chunk
	Blocks      BlockData // The cached block data for the chunk
	numVertices int32     // The number of vertices to render
	vao, vbo    uint32    // OpenGL buffers
}

// NewChunk creates a new, empty chunk with no block, rendering, or lighting
// data.
func newChunk(p, q int) *Chunk {
	// Create a VAO and VBO, but don't upload any data
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	return &Chunk{P: p, Q: q, vao: vao, vbo: vbo}
}

// Destroy releases all resources allocated when creating a chunk.
func (c *Chunk) destroy() {
	gl.DeleteBuffers(1, &c.vbo)
	gl.DeleteVertexArrays(1, &c.vao)
}

// Render the chunk to the screen.
func (c *Chunk) render(info RenderInfo) {
	// Don't bother rendering a chunk that's yet to be loaded, or has no vertex
	// data to render
	if c.Blocks == nil || c.numVertices == 0 {
		return
	}

	// Render the chunk
	gl.BindVertexArray(c.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, c.numVertices)
}

// BlockData represents an array of blocks within a chunk.
type BlockData []Block

// NewBlockData creates a new blocks array for a chunk, with length equal to
// the number of blocks in a chunk.
func newBlockData() BlockData {
	return make([]Block, ChunkWidth*ChunkHeight*ChunkDepth)
}

// At returns the block at the given coordinate within the block list. If the
// given coordinates are outside the block list's boundaries, then returns
func (b BlockData) At(x, y, z int) *Block {
	// Prevent an array out of bounds exception
	if x < 0 || x >= ChunkWidth ||
		y < 0 || y >= ChunkHeight ||
		z < 0 || z >= ChunkDepth {
		return nil
	}
	return &b[y*ChunkWidth*ChunkDepth+z*ChunkWidth+x]
}

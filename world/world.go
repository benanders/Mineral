package world

import (
	"log"
	"unsafe"

	"github.com/benanders/mineral/camera"
	"github.com/benanders/mineral/render"

	"github.com/chewxy/math32"
	"github.com/go-gl/gl/v3.3-core/gl"
)

const (
	// MaxRenderRadius is the maximum number of chunks ahead of the player which
	// we can feasibly render.
	MaxRenderRadius = 32
)

// ToWorldSpace returns the absolute coordinate of the block that contains the
// given world-space coordinate.
func ToWorldSpace(wx, wy, wz float32) (int, int, int) {
	return int(math32.Floor(wx)), int(math32.Floor(wy)), int(math32.Floor(wz))
}

// ToChunkSpace returns the coordinates of the chunk and the block within that
// chunk that contain the given world-space coordinate.
func ToChunkSpace(wx, wy, wz int) (p, q, x, y, z int) {
	// Use floor to always round down towards negative infinity. Otherwise the
	// 4 chunks around the centre of the world would have a (p, q) of (0, 0)
	p = int(math32.Floor(float32(wx) / float32(ChunkWidth)))
	q = int(math32.Floor(float32(wz) / float32(ChunkDepth)))

	// Go's modulus operator is stupid and returns negative numbers, so we fix
	// this by adding on `ChunkWidth` or `ChunkDepth` if necessary
	x = wx % ChunkWidth
	if x < 0 {
		x += ChunkWidth
	}
	y = wy
	z = wz % ChunkDepth
	if z < 0 {
		z += ChunkDepth
	}
	return
}

// World manages the loading, unloading, and rendering of chunks.
type World struct {
	RenderRadius uint                // Current render distance
	chunks       map[chunkPos]*Chunk // All loaded chunks
	loading      []chan interface{}  // Channels to goroutines loading chunks
	blocksInfo   BlocksInfo          // Information about each block type

	// Shader program uniforms and attributes
	program       uint32
	mvpUnf        int32
	blockAtlasUnf int32
	posAttr       uint32
	normalAttr    uint32
	uvAttr        uint32

	// Block texture atlas ID
	terrainTexture uint32
}

// New creates a new world instance with no loaded chunks.
func New(renderRadius uint) *World {
	// Load the chunk rendering program
	program, err := render.LoadShaders(
		"shaders/chunkVert.glsl",
		"shaders/chunkFrag.glsl")
	if err != nil {
		log.Fatalln(err)
	}
	gl.UseProgram(program)

	// Cache the uniform locations
	mvpUnf := gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	blockAtlasUnf := gl.GetUniformLocation(program, gl.Str("blockAtlas\x00"))

	// Cache the attribute locations
	posAttr := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	normalAttr := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
	uvAttr := uint32(gl.GetAttribLocation(program, gl.Str("uv\x00")))

	// Load information about each block type and create the block texture atlas
	blocksInfo, terrainTexture := loadBlocksInfo()

	return &World{
		renderRadius,
		make(map[chunkPos]*Chunk, 0),
		make([]chan interface{}, 0),
		blocksInfo,
		program, mvpUnf, blockAtlasUnf, posAttr, normalAttr, uvAttr,
		terrainTexture,
	}
}

// Destroy unloads all the currently loaded chunks.
func (w *World) Destroy() {
	gl.DeleteProgram(w.program)
	gl.DeleteTextures(1, &w.terrainTexture)

	// Close all loading to goroutines loading chunks
	for _, ch := range w.loading {
		close(ch)
	}

	// Destroy all loaded chunks
	for pos, chunk := range w.chunks {
		chunk.destroy()
		delete(w.chunks, pos)
	}
}

// FindChunk checks to see if the chunk at the given coordinates is already
// loaded, and if so returns a pointer to it. Otherwise, returns nil.
func (w *World) FindChunk(p, q int) *Chunk {
	if chunk, ok := w.chunks[chunkPos{p, q}]; ok {
		return chunk
	}
	return nil
}

// GetBlockInfo returns information about a particular block type.
func (w *World) GetBlockInfo(block Block) *BlockInfo {
	return w.blocksInfo.get(block)
}

// BlockVertexGenResult stores the block and vertex data generated for a chunk
// upon initially loading the chunk.
type blockVertexGenResult struct {
	p, q     int       // The location of the chunk we generated vertex data for
	blocks   BlockData // The generated block data
	vertices []float32 // The generated vertex data
}

// GenChunk first generates block data for a chunk, then the chunk's vertex
// data from this, on a separate goroutine.
//
// If the chunk at the given coordinates is already loaded, then the function
// does nothing.
func (w *World) GenChunk(p, q int) {
	// Check the chunk isn't already loaded
	if chunk := w.FindChunk(p, q); chunk != nil {
		return
	}

	// Load the chunk's block and vertex data
	ch := make(chan interface{})
	w.loading = append(w.loading, ch)
	go (func() {
		blocks := genBlocks(p, q)
		vertices := genVertices(vertexGenInfo{p, q, blocks, &w.blocksInfo})
		ch <- blockVertexGenResult{p, q, blocks, vertices}
	})()
}

// VertexGenResult stores the data generated when a chunk's vertex data is
// reloaded from its existing block data.
type vertexGenResult struct {
	p, q     int       // The location of the chunk we generated vertex data for
	vertices []float32 // The generated vertex data itself
}

// RegenChunk regenerates the vertex data for the chunk at the given
// coordinates on a separate goroutine, using its existing block data. This
// should be called if the chunk's block data is modified (e.g. after placing a
// new block).
//
// If the chunk at the given coordinates isn't already loaded, then the function
// does nothing.
func (w *World) regenChunk(p, q int) {
	// Check that the chunk loaded, bailing if it isn't
	chunk := w.FindChunk(p, q)
	if chunk == nil || chunk.Blocks == nil {
		return
	}

	// Copy block data into a new array, in case the chunk is unloaded while
	// we're in the middle of loading it
	copied := newBlockData()
	copy(copied, chunk.Blocks)

	// Load the vertex data on a separate goroutine
	ch := make(chan interface{})
	w.loading = append(w.loading, ch)
	go (func() {
		vertices := genVertices(vertexGenInfo{p, q, copied, &w.blocksInfo})
		ch <- vertexGenResult{p, q, vertices}
	})()
}

// Update is called every update tick, and checks to see if any loading tasks
// are finished.
func (w *World) Update() {
	// Select across all channels
	for _, ch := range w.loading {
		select {
		case result := <-ch:
			w.handleFinishedTask(result)
		default: // We want non-blocking channel reads
		}
	}
}

// HandleFinishedTask takes the data generated by a chunk loading task and
// updates the relevant chunk with the information.
func (w *World) handleFinishedTask(result interface{}) {
	switch r := result.(type) {
	case blockVertexGenResult:
		// Loaded all information to do with a chunk
		chunk := newChunk(r.p, r.q)
		chunk.Blocks = r.blocks
		w.uploadChunk(chunk, r.vertices)
		w.chunks[chunkPos{r.p, r.q}] = chunk
	case vertexGenResult:
		// Reloaded a chunk's vertex data
		chunk := w.FindChunk(r.p, r.q)
		if chunk == nil {
			// Chunk was unloaded while we were loading its data; do nothing
			return
		}
		w.uploadChunk(chunk, r.vertices)
	}
}

// UploadChunk pushes the new vertex data for a chunk to the GPU.
func (w *World) uploadChunk(chunk *Chunk, vertices []float32) {
	chunk.numVertices = int32(len(vertices)) / valuesPerVertex

	// Upload the vertex data by deleting the current vertex buffer and
	// reallocating it
	gl.BindVertexArray(chunk.vao)
	gl.DeleteBuffers(1, &chunk.vbo)
	gl.GenBuffers(1, &chunk.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, chunk.vbo)
	var ptr unsafe.Pointer
	if len(vertices) > 0 {
		ptr = gl.Ptr(vertices)
	}

	// For some reason (I have no idea why, maybe something to do with Go's
	// internal representation of slices, and how they have a length/capacity
	// value associated with them in a struct??) we need to add 1 to the length
	// of the slice that we're copying to the GPU. If we don't do this, the
	// last value in the vertex data is cut off.
	gl.BufferData(gl.ARRAY_BUFFER, (len(vertices)+1)*4, ptr, gl.STATIC_DRAW)

	// Set the vertex attributes on the new buffer
	gl.UseProgram(w.program)

	// Position attribute
	gl.EnableVertexAttribArray(w.posAttr)
	gl.VertexAttribPointer(w.posAttr, 3, gl.FLOAT, false, valuesPerVertex*4,
		gl.PtrOffset(0))

	// Normal attribute
	gl.EnableVertexAttribArray(w.normalAttr)
	gl.VertexAttribPointer(w.normalAttr, 3, gl.FLOAT, false, valuesPerVertex*4,
		gl.PtrOffset(3*4))

	// UV attribute
	gl.EnableVertexAttribArray(w.uvAttr)
	gl.VertexAttribPointer(w.uvAttr, 3, gl.FLOAT, false, valuesPerVertex*4,
		gl.PtrOffset(6*4))
}

// RenderInfo stores information required by the world for rendering.
type RenderInfo struct {
	Camera *camera.Camera
}

// Render draws all loaded chunks with vertex data to the screen.
func (w *World) Render(info RenderInfo) {
	// Enable some OpenGL state
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	// Use the chunk shader program and set uniforms
	gl.UseProgram(w.program)
	gl.UniformMatrix4fv(w.mvpUnf, 1, false, &info.Camera.View[0])
	gl.Uniform1i(w.blockAtlasUnf, blockAtlasSlot)

	// Render each chunk
	for _, chunk := range w.chunks {
		chunk.render(info)
	}

	// Reset the OpenGL state
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
}

package world

import (
	"log"
	"unsafe"

	"github.com/benanders/mineral/asset"
	"github.com/benanders/mineral/block"
	"github.com/benanders/mineral/camera"
	"github.com/benanders/mineral/render"
	"github.com/chewxy/math32"

	"github.com/go-gl/gl/v3.3-core/gl"
)

const (
	// MaxRenderRadius is the maximum value for the render radius of chunks. It
	// is the maximum distance a player could see ahead of them (measured in
	// chunks).
	MaxRenderRadius = 32

	// TerrainTextureSlot is the OpenGL texture slot that the terrain texture
	// is loaded into.
	terrainTextureSlot = 0
)

// World stores all the loaded chunks and loads/unloads chunks as required.
type World struct {
	RenderRadius uint                // Current render distance
	chunks       map[ChunkPos]*Chunk // All loaded chunks
	channels     []chan interface{}  // Channels to goroutines loading chunks

	program           uint32
	mvpUnf            int32
	terrainTextureUnf int32
	posAttr           uint32
	normalAttr        uint32
	uvAttr            uint32

	terrainTexture uint32
}

// NewWorld creates a new world instance with no loaded chunks yet.
func New(renderRadius uint) *World {
	// Load the block variants from the asset files
	block.LoadVariants()

	// Get the chunk vertex and fragment shaders
	vert, err := asset.Asset("shaders/chunkVert.glsl")
	if err != nil {
		log.Fatalln("failed to load shaders/chunkVert.glsl: ", err)
	}
	frag, err := asset.Asset("shaders/chunkFrag.glsl")
	if err != nil {
		log.Fatalln("failed to load shaders/chunkFrag.glsl: ", err)
	}

	// Create the chunk rendering program
	program, err := render.LoadShaders(string(vert), string(frag))
	if err != nil {
		log.Fatalln("failed to load chunk shader: ", err)
	}
	gl.UseProgram(program)

	// Cache the uniform locations
	mvpUnf := gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	terrainTextureUnf := gl.GetUniformLocation(program, gl.Str("terrain\x00"))

	// Cache the attribute locations
	posAttr := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	normalAttr := uint32(gl.GetAttribLocation(program, gl.Str("normal\x00")))
	uvAttr := uint32(gl.GetAttribLocation(program, gl.Str("uv\x00")))

	// Load the terrain texture atlas
	terrainTexture := block.LoadTerrainAtlas(terrainTextureSlot)

	return &World{
		renderRadius,
		make(map[ChunkPos]*Chunk, 0),
		make([]chan interface{}, 0),
		program, mvpUnf, terrainTextureUnf, posAttr, normalAttr, uvAttr,
		terrainTexture,
	}
}

// Destroy unloads all the currently loaded chunks.
func (w *World) Destroy() {
	gl.DeleteProgram(w.program)
	gl.DeleteTextures(1, &w.terrainTexture)

	// Destroy all loaded chunks
	for _, chunk := range w.chunks {
		chunk.Destroy()
	}
	w.chunks = nil

	// Close all pending channels
	for _, ch := range w.channels {
		close(ch)
	}
}

// Gets the coordinates of the chunk and block that contain the given world
// coordinate.
func Chunked(wx, wy, wz int) (p, q, x, y, z int) {
	// Use floor to always round down towards negative infinity. Otherwise the
	// 4 chunks around the centre of the world would have a (p, q) of (0, 0)
	p = int(math32.Floor(float32(wx) / float32(block.ChunkWidth)))
	q = int(math32.Floor(float32(wz) / float32(block.ChunkDepth)))

	// Go's modulo operator is stupid and returns negative numbers, so we fix
	// this by adding on `ChunkWidth` or `ChunkDepth`
	x = wx % block.ChunkWidth
	if x < 0 {
		x += block.ChunkWidth
	}
	y = wy
	z = wz % block.ChunkDepth
	if z < 0 {
		z += block.ChunkDepth
	}
	return
}

// VertexLoadResult stores the data generated when a chunk's vertex data is
// reloaded from its existing block data.
type vertexLoadResult struct {
	p, q     int
	vertices []float32
}

// ReloadChunk queues the chunk at the given coordinates for a vertex data
// reload. If the chunk isn't yet loaded, then does nothing (doesn't load it).
func (w *World) reloadChunk(p, q int) {
	// Find the chunk
	chunk := w.FindChunk(p, q)
	if chunk == nil || chunk.Blocks == nil {
		return // Chunk isn't loaded
	}

	// Copy block data into a new array, in case the chunk is unloaded while
	// we're in the middle of loading it
	copiedBlocks := newBlockData()
	copy(copiedBlocks, chunk.Blocks)

	// Load the vertex data on
	ch := make(chan interface{})
	w.channels = append(w.channels, ch)
	go (func() {
		vertices := genVertices(p, q, copiedBlocks)
		ch <- vertexLoadResult{p, q, vertices}
	})()
}

// CompleteLoadResult stores the data generated when a chunk's block, vertex,
// and lighting data is all loaded at once.
type completeLoadResult struct {
	p, q     int
	blocks   BlockData
	vertices []float32
}

// LoadChunk queues the chunk at the given coordinates for loading. If the
// chunk is already loaded, then does nothing (doesn't reload its data).
func (w *World) LoadChunk(p, q int) {
	// Check the chunk isn't already loaded
	if chunk := w.FindChunk(p, q); chunk != nil {
		return
	}

	// Load the chunk's block and vertex data
	ch := make(chan interface{})
	w.channels = append(w.channels, ch)
	go (func() {
		blocks := genBlocks(p, q)
		vertices := genVertices(p, q, blocks)
		ch <- completeLoadResult{p, q, blocks, vertices}
	})()
}

// Update should be called every update tick to check for completed load tasks.
func (w *World) Update() {
	// Select across all chunk loading channels
	for _, ch := range w.channels {
		select {
		case result := <-ch:
			w.handleFinishedTask(result)
		default:
			// We want non-blocking channel reads
		}
	}
}

// HandleFinishedTask takes the data generated by a chunk loading task and
// updates the relevant chunk with the information.
func (w *World) handleFinishedTask(result interface{}) {
	switch r := result.(type) {
	case completeLoadResult:
		// Loaded all information to do with a chunk
		chunk := newChunk(r.p, r.q)
		chunk.Blocks = r.blocks
		w.uploadChunk(chunk, r.vertices)
		w.chunks[ChunkPos{r.p, r.q}] = chunk
	case vertexLoadResult:
		// Reloaded a chunk's vertex data
		chunk := w.FindChunk(r.p, r.q)
		if chunk == nil {
			// Chunk was unloaded while we were loading its data; do nothing
			return
		}
		w.uploadChunk(chunk, r.vertices)
	}
}

// UploadChunk pushes the new vertex data for a chunk onto the GPU.
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

// FindChunk checks to see if the chunk at the given coordinates is already
// loaded, and if so returns a pointer to it. Otherwise, returns nil.
func (w *World) FindChunk(p, q int) *Chunk {
	if chunk, ok := w.chunks[ChunkPos{p, q}]; ok {
		return chunk
	}
	return nil
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
	gl.Uniform1i(w.terrainTextureUnf, terrainTextureSlot)

	// Render each chunk
	for _, chunk := range w.chunks {
		chunk.render(info)
	}

	// Reset the OpenGL state
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
}

// ChunkPos stores the position of a chunk.
type ChunkPos struct {
	p, q int
}

// Chunk stores information associated with a chunk, including OpenGL rendering
// information, block data, vertex data, and lighting data.
type Chunk struct {
	P, Q        int       // The position of the chunk, in chunk coordinates
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
func (c *Chunk) Destroy() {
	gl.DeleteBuffers(1, &c.vbo)
	gl.DeleteVertexArrays(1, &c.vao)
}

// Render the chunk to the screen.
func (c *Chunk) render(info RenderInfo) {
	// Don't bother rendering an unloaded chunk
	if c.Blocks == nil {
		return
	}

	// Render the chunk
	gl.BindVertexArray(c.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, c.numVertices)
}

// BlockData represents an array of blocks within a chunk.
type BlockData []block.Block

// NewBlockData creates a new blocks array for a chunk, with length equal to
// the number of blocks within a chunk.
func newBlockData() BlockData {
	return make([]block.Block,
		block.ChunkWidth*block.ChunkHeight*block.ChunkDepth)
}

// At returns the block at the given coordinate within the block list. If the
// given coordinates are outside the block list's boundaries, then returns
func (b BlockData) At(x, y, z int) *block.Block {
	if x < 0 || x >= block.ChunkWidth ||
		y < 0 || y >= block.ChunkHeight ||
		z < 0 || z >= block.ChunkDepth {
		return nil
	} else {
		return &b[y*block.ChunkWidth*block.ChunkDepth+z*block.ChunkWidth+x]
	}
}

package world

// ValuesPerVertex tells us the number of floating point values emitted per
// vertex.
const valuesPerVertex = 8

// VertexGenInfo contains the necessary information to generate vertex data for
// a chunk.
type vertexGenInfo struct {
	p, q   int       // The chunk to generate vertex data for
	blocks BlockData // A copy of the chunk's block data

	// Information about each block type, indexed by ID. This is only ever read
	// from (never written to), so we're not going to get any race conditions.
	blocksInfo *BlocksInfo
}

// GenVertices takes the block data for a chunk and generates the chunk's
// vertex data, based on the faces of the blocks that are visible.
func genVertices(info vertexGenInfo) []float32 {
	// Generate vertex data for each block in the chunk
	var vertices []float32
	for x := 0; x < ChunkWidth; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkDepth; z++ {
				genVerticesForBlock(&vertices, info, x, y, z)
			}
		}
	}

	return vertices
}

// GenVerticesForBlock determines which faces of the block at the given
// coordinates are visible, and adds them to the vertex data.
func genVerticesForBlock(vertices *[]float32, info vertexGenInfo, x, y, z int) {
	// Don't generate vertices for invisible blocks
	current := info.blocks.At(x, y, z)
	if current == nil || !info.blocksInfo.get(*current).Visible {
		return
	}

	// Generate vertex data for each face
	for face := FaceLeft; face <= FaceBack; face++ {
		// Get the coordinate of the block next to this face
		nx, ny, nz := face.normal()
		bx, by, bz := x+nx, y+ny, z+nz

		// Only generate vertex data if the block next to this face is
		// semi-transparent, or if the block is at a chunk border
		neighbour := info.blocks.At(bx, by, bz)
		if neighbour == nil || info.blocksInfo.get(*neighbour).Transparent {
			genVerticesForFace(vertices, info, *current, x, y, z, face)
		}
	}
}

// GenVerticesForFace adds the vertex data for a visible face of a block to
// the vertices list.
func genVerticesForFace(vertices *[]float32, info vertexGenInfo, block Block,
	x, y, z int, face blockFace) {
	// All vertices that make up a cube
	cubeVertices := [...][3]float32{
		{0.0, 0.0, 1.0}, // Left,  bottom, front
		{1.0, 0.0, 1.0}, // Right, bottom, front
		{1.0, 1.0, 1.0}, // Right, top,    front
		{0.0, 1.0, 1.0}, // Left,  top,    front
		{0.0, 0.0, 0.0}, // Left,  bottom, back
		{1.0, 0.0, 0.0}, // Right, bottom, back
		{1.0, 1.0, 0.0}, // Right, top,    back
		{0.0, 1.0, 0.0}, // Left,  top,    back
	}

	// Vertices that make up each face of a cube. A vertex is specified by an
	// index into the `cubeVertices` array above
	faceIndices := [...][6]uint16{
		{7, 4, 0, 0, 3, 7}, // Left
		{2, 1, 5, 5, 6, 2}, // Right
		{6, 7, 3, 3, 2, 6}, // Top
		{0, 4, 5, 5, 1, 0}, // Bottom
		{3, 0, 1, 1, 2, 3}, // Front
		{6, 5, 4, 4, 7, 6}, // Back
	}

	// UVs for each vertex, for all faces on a cube
	faceUVs := [...][2]float32{
		{0.0, 0.0}, {0.0, 1.0}, {1.0, 1.0}, {1.0, 1.0}, {1.0, 0.0}, {0.0, 0.0},
	}

	// Iterate over the 6 vertices of the 2 triangles that make up the face
	for vertex := 0; vertex < 6; vertex++ {
		// Position
		position := &cubeVertices[faceIndices[face][vertex]]
		*vertices = append(*vertices, float32(x)+position[0])
		*vertices = append(*vertices, float32(y)+position[1])
		*vertices = append(*vertices, float32(z)+position[2])

		// Normal
		nx, ny, nz := face.normal()
		*vertices = append(*vertices, float32(nx))
		*vertices = append(*vertices, float32(ny))
		*vertices = append(*vertices, float32(nz))

		// UV
		uv := info.blocksInfo.get(block).UV
		w, h := uv.Size()
		*vertices = append(*vertices, uv.X+w*faceUVs[vertex][0])
		*vertices = append(*vertices, uv.Y+h*faceUVs[vertex][1])
	}
}

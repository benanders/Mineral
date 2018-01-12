package world

// ValuesPerVertex tells us the number of floating point numbers emitted per
// vertex within the vertex data.
const valuesPerVertex = 8

// GenVertices takes the block data for a chunk and generates the chunk's
// vertex data, based on the faces of the blocks that are visible.
func genVertices(p, q int, blocks BlockData) []float32 {
	// Create the vertices array
	vertices := make([]float32, 0)

	// Generate vertex data for each block in the chunk
	for x := 0; x < ChunkWidth; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkDepth; z++ {
				genVerticesForBlock(&vertices, blocks, x, y, z)
			}
		}
	}

	return vertices
}

// GenVerticesForBlock determines which faces of the block at the given
// coordinates are visible, and adds them to the vertex data.
func genVerticesForBlock(vertices *[]float32, blocks BlockData, x, y, z int) {
	// Don't generate vertices for air
	if block := blocks.At(x, y, z); *block == BlockAir {
		return
	}

	// Generate vertex data for each face
	for face := faceLeft; face <= faceBack; face++ {
		// Get the coordinate of the block next to this face
		nx, ny, nz := face.normal()
		bx, by, bz := x+nx, y+ny, z+nz

		// Only generate vertex data if the block next to this face is
		// semi-transparent
		if blocks.At(bx, by, bz).IsTransparent() {
			genVerticesForFace(vertices, blocks, x, y, z, face)
		}
	}
}

// GenVerticesForFace adds the vertex data for a visible face of a block to
// the vertices list.
func genVerticesForFace(vertices *[]float32, blocks BlockData, x, y, z int,
	face blockFace) {
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

	// Iterate over the 6 vertices of the face
	for vertex := 0; vertex < 6; vertex++ {
		// Position
		faceVertices := &cubeVertices[faceIndices[face][vertex]]
		*vertices = append(*vertices, float32(x)+faceVertices[0])
		*vertices = append(*vertices, float32(y)+faceVertices[1])
		*vertices = append(*vertices, float32(z)+faceVertices[2])

		// Normal
		nx, ny, nz := face.normal()
		*vertices = append(*vertices, float32(nx))
		*vertices = append(*vertices, float32(ny))
		*vertices = append(*vertices, float32(nz))

		// UV
		*vertices = append(*vertices, faceUVs[vertex][0])
		*vertices = append(*vertices, faceUVs[vertex][1])
	}
}

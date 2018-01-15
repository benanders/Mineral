package block

import "image"

// TerrainAtlas represents the terrain texture atlas image, compsed from
// a series of smaller images into a larger texture upon load.
type TerrainAtlas struct {
	img     *image.RGBA // The underlying texture pixel data
	Texture uint32      // The OpenGL texture ID
	Slot    uint32      // The slot the OpenGL texture is loaded into
}

// NewTerrainAtlas creates a new terrain atlas from a series of block variants
// and the image files that they reference.
func NewTerrainAtlas(variants []BlockVariant) TerrainAtlas {

}

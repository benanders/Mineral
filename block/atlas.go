package block

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/png" // Required to load png files
	"log"

	"github.com/benanders/mineral/asset"
	"github.com/benanders/mineral/render"
)

const (
	// The size of each block texture, in pixels.
	BlockTextureWidth  = 16
	BlockTextureHeight = 16

	// The size of the terrain atlas image, in pixels.
	AtlasTextureWidth  = 256
	AtlasTextureHeight = 256
)

// BlockUVs contains the coorindate of the 16x16 pixel block in the texture
// atlas that corresponds to each block variant. It is indexed by block ID.
var blockUVs []BlockUV

// BlockUV stores the coordinate of the 16x16 pixel block in the texture atlas
// to use to render all 6 faces of a block variant.
type BlockUV struct {
	X, Y float32
}

// Size returns the size of a block texture in the texture atlas, scaled such
// that a size of (1.0, 1.0) represents the entire texture atlas. The size is
// used to calculate the UV coordinates passed to OpenGL for the block texture.
func (uv BlockUV) Size() (float32, float32) {
	return float32(BlockTextureWidth) / float32(AtlasTextureWidth),
		float32(BlockTextureHeight) / float32(AtlasTextureHeight)
}

// LoadTerrainAtlas creates a new terrain atlas image from the individual
// textures for each block, uploads it to the GPU in the given texture slot,
// and returns an OpenGL texture ID. It also populates the blockUVs array with
// UV data for each block variant.
func LoadTerrainAtlas(slot uint32) uint32 {
	// Create the terrain atlas image, 256x256 pixels (16x16 blocks)
	rect := image.Rect(0, 0, AtlasTextureWidth, AtlasTextureHeight)
	atlasImg := image.NewRGBA(rect)

	// Load each png and place it into the atlas
	x, y := 0, 0
	for _, variant := range blockVariants {
		// Only bother getting an image if the block is visible
		if !variant.Visible {
			// Still create an empty entry in the block UVs list so we can
			// index the array by block ID
			blockUVs = append(blockUVs, BlockUV{0, 0})
			continue
		}

		// Get the image
		pngData, err := asset.Asset(variant.Texture)
		if err != nil {
			log.Fatalln("failed to load image `" + variant.Texture +
				"` for block " + variant.Name)
		}

		// Load the pixel data from the image
		blockImg, _, err := image.Decode(bytes.NewReader(pngData))
		if err != nil {
			log.Fatalln("failed to load png for block " + variant.Name)
		}

		// Add the data to the atlas
		srcPoint := image.Point{0, 0}
		dstRect := image.Rect(x, y, x+BlockTextureWidth, y+BlockTextureHeight)
		draw.Draw(atlasImg, dstRect, blockImg, srcPoint, draw.Over)

		// Add the coordinate to the block UVs list
		xRel := float32(x) / float32(AtlasTextureWidth)
		yRel := float32(y) / float32(AtlasTextureHeight)
		blockUVs = append(blockUVs, BlockUV{xRel, yRel})

		// Increment the offset in the atlas
		x += BlockTextureWidth
		if x > AtlasTextureWidth-BlockTextureWidth {
			x = 0
			y += BlockTextureHeight
		}

		// Check we're still within the image bounds
		if y > AtlasTextureHeight-BlockTextureHeight {
			log.Fatalln("failed to fit all blocks in terrain texture atlas")
		}
	}

	// Upload the texture to the GPU
	return render.LoadTexture(atlasImg, slot)
}

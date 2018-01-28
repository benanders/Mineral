package world

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/png" // Block textures are provided as .png images
	"log"
	"sort"

	"github.com/benanders/mineral/asset"
	"github.com/benanders/mineral/math"
	"github.com/benanders/mineral/render"

	"github.com/BurntSushi/toml"
	"github.com/go-gl/mathgl/mgl32"
)

// Block is an ID representing the type of a block within the world.
type Block uint32

// BlockFace represents one of the 6 faces of a block.
type blockFace uint

// All block faces.
const (
	FaceLeft blockFace = iota
	FaceRight
	FaceTop
	FaceBottom
	FaceFront
	FaceBack
)

// FaceNormals is an array indexed by block face that tells us the normal
// vector for each face.
var faceNormals = [...][3]int{
	{-1, 0, 0}, // Left
	{1, 0, 0},  // Right
	{0, 1, 0},  // Top
	{0, -1, 0}, // Bottom
	{0, 0, 1},  // Front
	{0, 0, -1}, // Back
}

// Normal tells us the normal vector for a face.
func (f blockFace) normal() (int, int, int) {
	return faceNormals[f][0], faceNormals[f][1], faceNormals[f][2]
}

const (
	// BlockAtlasSlot is the OpenGL texture slot into which the block atlas
	// image is to be loaded.
	blockAtlasSlot = 0

	// The size of each block texture, in pixels.
	blockTextureWidth  = 16
	blockTextureHeight = 16

	// The size of the block atlas image, in pixels.
	atlasTextureWidth  = 256
	atlasTextureHeight = 256
)

// BlocksInfo contains the properties of every block type.
type BlocksInfo []*BlockInfo

// Get returns information for the given block type.
func (info *BlocksInfo) get(b Block) *BlockInfo {
	return (*info)[b]
}

// BlockInfo contains the properties of a block type.
type BlockInfo struct {
	Name        string // Display name of the block
	Visible     bool   // True if the block actually renders something
	Collidable  bool   // True if the block has a collidable AABB
	Transparent bool   // True if we can see the block behind at any angle
	Texture     string // Path to the texture to use for all faces
	UV          FaceUV // UV coordinates for each face
}

// AABB returns an axis aligned bounding box for the block, used for collision
// detection.
func (info *BlockInfo) AABB(p, q, x, y, z int) math.AABB {
	// Add 0.5 since the AABB struct requires we specify the centre of the
	// block, and blocks are always 1x1 units
	rx := float32(p*ChunkWidth+x) + 0.5
	ry := float32(y) + 0.5
	rz := float32(q*ChunkDepth+z) + 0.5
	return math.AABB{
		Center: mgl32.Vec3{rx, ry, rz},
		Size:   mgl32.Vec3{1.0, 1.0, 1.0},
	}
}

// FaceUV represents the base UV coordinate for a block face in the block
// texture atlas.
type FaceUV struct {
	X, Y float32
}

// Size returns the size of a block texture in the texture atlas, scaled such
// that a size of (1.0, 1.0) represents the entire texture atlas. The size is
// used to calculate the UV coordinates passed to OpenGL for the block texture.
func (uv FaceUV) Size() (float32, float32) {
	return float32(blockTextureWidth) / float32(atlasTextureWidth),
		float32(blockTextureHeight) / float32(atlasTextureHeight)
}

// LoadBlocksInfo reads the properties of every block from the asset files and
// constructs the texture atlas.
//
// Returns an array, indexed by block ID, of information for each block type,
// and the OpenGL ID for the block texture atlas.
func loadBlocksInfo() (BlocksInfo, uint32) {
	blocksInfo := loadBlocksProperties()
	blockAtlas := loadBlockAtlas(blockAtlasSlot, blocksInfo)
	return blocksInfo, blockAtlas
}

// LoadBlocksProperties reads the properties of every block in the world from
// the asset files.
func loadBlocksProperties() BlocksInfo {
	// Get the file name of every block
	blockNames, err := asset.AssetDir("blocks")
	if err != nil {
		log.Fatalln("asset/data/blocks not found")
	}

	// Sort in alphabetical order, so that block IDs are consistent every time
	// the game is launched
	sort.Strings(blockNames)

	// Load information for each block
	blocksInfo := make([]*BlockInfo, 0)
	for _, blockName := range blockNames {
		// Get the TOML source
		source, err := asset.Asset("blocks/" + blockName)
		if err != nil {
			log.Fatalln("failed to load "+blockName+": ", err)
		}

		// Decode the TOML source
		var info BlockInfo
		_, err = toml.Decode(string(source), &info)
		if err != nil {
			log.Fatalln("failed to decode "+blockName+": ", err)
		}

		// Add to the list of block variants
		blocksInfo = append(blocksInfo, &info)
	}

	return blocksInfo
}

// LoadBlockAtlas creates a new texture atlas image from the individual textures
// for each block, uploads it to the GPU in the given texture slot, and returns
// an OpenGL texture ID.
//
// The function sets the UV coordinates for each block type in the blockInfos
// array.
func loadBlockAtlas(slot uint32, blocksInfo BlocksInfo) uint32 {
	// Create the block atlas image
	rect := image.Rect(0, 0, atlasTextureWidth, atlasTextureHeight)
	atlasImg := image.NewRGBA(rect)

	// Load each png and place it into the atlas
	x, y := 0, 0
	for _, info := range blocksInfo {
		// Only bother getting an image if the block is visible
		if !info.Visible {
			continue
		}

		// Check we've still got enough room in the atlas to fit another texture
		if y > atlasTextureHeight-blockTextureHeight {
			log.Fatalln("failed to fit all block textures in block atlas")
		}

		// Get the .png file that contains the block's texture
		pngData, err := asset.Asset(info.Texture)
		if err != nil {
			log.Fatalln("failed to load image `" + info.Texture +
				"` for block " + info.Name)
		}

		// Decode the .png file
		blockImg, _, err := image.Decode(bytes.NewReader(pngData))
		if err != nil {
			log.Fatalln("failed to decode png image `" + info.Texture +
				"` for block " + info.Name)
		}

		// Ensure the block texture is of the correct size
		w := blockImg.Bounds().Max.X - blockImg.Bounds().Min.X
		h := blockImg.Bounds().Max.Y - blockImg.Bounds().Min.Y
		if w != blockTextureWidth || h != blockTextureHeight {
			log.Fatalln("image for block " + info.Name + " is incorrect size")
		}

		// Copy the block's texture into the texture atlas
		srcPoint := image.Point{0, 0}
		dstRect := image.Rect(x, y, x+w, y+h)
		draw.Draw(atlasImg, dstRect, blockImg, srcPoint, draw.Over)

		// Set the block's UV coordinates
		info.UV.X = float32(x) / float32(atlasTextureWidth)
		info.UV.Y = float32(y) / float32(atlasTextureHeight)

		// Increment the offset at which textures are placed in the atlas
		x += blockTextureWidth
		if x > atlasTextureWidth-blockTextureWidth {
			x = 0
			y += blockTextureHeight
		}
	}

	// Upload the texture to the GPU
	return render.LoadTexture(atlasImg, slot)
}

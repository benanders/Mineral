package block

import (
	"log"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/benanders/mineral/asset"
)

// BlockVariant contains the properties of a block type.
type blockVariant struct {
	Name        string // Display name of the block
	Visible     bool   // True if the block actually renders something
	Collidable  bool   // True if the block has a collidable AABB
	Transparent bool   // True if we can see the block behind at any angle
	Texture     string // Path to the texture to use for all faces
}

// BlockVariants contains a list of properties for every block in the game.
var blockVariants []blockVariant

// LoadVariants reads the properties of every block in the world from the asset
// files.
func LoadVariants() {
	// Get the file name of every block
	blockNames, err := asset.AssetDir("blocks")
	if err != nil {
		log.Fatalln("asset/data/blocks not found")
	}

	// Sort in alphabetical order
	sort.Strings(blockNames)

	// Load information for each block
	for _, blockName := range blockNames {
		// Get the TOML source
		source, err := asset.Asset("blocks/" + blockName)
		if err != nil {
			log.Fatalln("failed to load "+blockName+": ", err)
		}

		// Decode the TOML source
		var variant blockVariant
		_, err = toml.Decode(string(source), &variant)
		if err != nil {
			log.Fatalln("failed to decode "+blockName+": ", err)
		}

		// Add to the list of block variants
		blockVariants = append(blockVariants, variant)
	}
}

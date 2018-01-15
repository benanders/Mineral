package main

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// SupportedVersions lists the supported versions of Minecraft that we can
// extract assets from. The versions are listed in order of preference, from
// most preferred to least preferred.
var supportedVersions = []string{
	"1.12.2",
	"1.12.1",
	"1.12",
	"1.11.2",
	"1.11.1",
	"1.11", // resource pack format changes before this version
}

// AssetMap specifies which assets from the original Minecraft game are to be
// copied across into Mineral's assets folder.
//
// All file paths for the Minecraft assets are specified relative to the
// `assets/minecraft` directory contained in the Minecraft `.jar` file.
//
// All output file paths are relative to the `assets/data` folder contained in
// this repository.
var assetMap = map[string]string{
	"assets/minecraft/textures/blocks/bedrock.png": "textures/blocks/bedrock.png",
	"assets/minecraft/textures/blocks/stone.png":   "textures/blocks/stone.png",
	"assets/minecraft/textures/blocks/dirt.png":    "textures/blocks/dirt.png",
}

func main() {
	// Locate the `minecraft` folder based on the operating system
	var minecraftFolder string
	if runtime.GOOS == "windows" {
		// Under `%APPDATA%\minecraft` on Windows
		if val, ok := os.LookupEnv("APPDATA"); ok {
			minecraftFolder = path.Join(val, "minecraft")
		} else {
			log.Fatalln("%APPDATA% environment variable not set")
		}
	} else if runtime.GOOS == "darwin" {
		// Under `~/Library/Application Support/minecraft` on macOS
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalln("failed to get home directory: ", err)
		}

		minecraftFolder = path.Join(home, "Library", "Application Support",
			"minecraft")
	} else {
		// Don't support anything else
		log.Println("unsupported operating system: " + runtime.GOOS)
		log.Fatalln("only macOS and windows are supported")
	}

	// Find the latest supported version in the `versions` folder
	versionsFolder := path.Join(minecraftFolder, "versions")
	version, ok := "", false
	for _, candidate := range supportedVersions {
		versionPath := path.Join(versionsFolder, candidate)
		if _, err := os.Stat(versionPath); err == nil {
			// Use this version
			version, ok = candidate, true
			break
		}
	}

	// Check we found a supported version
	if !ok {
		supportedVersionsString := strings.Join(supportedVersions, ", ")
		log.Println("no supported minecraft version installed")
		log.Fatalln("supported versions are: " + supportedVersionsString)
	}

	// Open a zip reader at the latest .jar file
	jarPath := path.Join(versionsFolder, version, version+".jar")
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		log.Fatalln("failed to read jar file at `"+jarPath+"`: ", err)
	}
	defer r.Close()

	// Get the path to the local assets folder, based off the executable path
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalln("failed to get current working directory")
	}
	assetsPath := path.Join(workingDir, "assets", "data")

	// Iterate over all the files in the zip
	count := 0
	for _, file := range r.File {
		// Check if the name matches an asset we're to copy across. The zip
		// spec says all file names in the zip are to use `/` for separating
		// directories.
		if copyPath, ok := assetMap[file.Name]; ok {
			count++

			// Open the file
			inputReader, err := file.Open()
			if err != nil {
				log.Fatalln("failed to open file in zip: " + file.Name)
			}

			// Read the entire file
			bytes, err := ioutil.ReadAll(inputReader)
			if err != nil {
				log.Fatalln("failed to read file in zip: " + file.Name)
			}
			inputReader.Close()

			// Open the output file
			splitCopyPath := strings.Split(copyPath, "/")
			outputPath := path.Join(assetsPath, path.Join(splitCopyPath...))
			os.MkdirAll(path.Dir(outputPath), 0700)
			outputWriter, err := os.Create(outputPath)
			if err != nil {
				log.Fatalln("failed to open output file: " + outputPath)
			}

			// Write to the output file
			_, err = outputWriter.Write(bytes)
			if err != nil {
				log.Fatalln("failed to write to output file: " + outputPath)
			}
			outputWriter.Close()
		}
	}

	// Check that we copied across all asset files
	if count != len(assetMap) {
		log.Fatalln("missing asset files! copied " + strconv.Itoa(count) +
			" files, expected " + strconv.Itoa(len(assetMap)) + " files")
	}

	log.Println("successfully copied " + strconv.Itoa(count) + " assets!")
}

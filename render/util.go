package render

import (
	"fmt"
	"image"
	"strings"

	"github.com/benanders/mineral/asset"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// LoadShaders compiles a vertex and fragment shader from an asset, creates a
// new OpenGL shader program, attaches the two shaders, and links the program.
func LoadShaders(vertexPath, fragmentPath string) (uint32, error) {
	// Get the source code for the shaders
	vertexSource, err := asset.Asset(vertexPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load asset `%v`: %v", vertexPath, err)
	}
	fragmentSource, err := asset.Asset(fragmentPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load asset `%v`: %v", vertexPath, err)
	}

	// Compile the vertex and fragment shaders
	vertex, err := compileShader(gl.VERTEX_SHADER, string(vertexSource))
	if err != nil {
		return 0, fmt.Errorf("failed to compile vertex shader `%v`: %v",
			vertexPath, err)
	}
	fragment, err := compileShader(gl.FRAGMENT_SHADER, string(fragmentSource))
	if err != nil {
		return 0, fmt.Errorf("failed to compile fragment shader `%v`: %v",
			fragmentPath, err)
	}

	// Create the program and attach the vertex and fragment shaders
	program := gl.CreateProgram()
	gl.AttachShader(program, vertex)
	gl.AttachShader(program, fragment)

	// Link the program
	err = linkProgram(program)
	if err != nil {
		return 0, fmt.Errorf("failed to link program (`%v` and `%v`): %v",
			vertexPath, fragmentPath, err)
	}

	return program, nil
}

// LoadShader compiles a shader from a string, checking for any compilation
// errors.
func compileShader(kind uint32, source string) (uint32, error) {
	// Create the shader and attach the source code to it
	shader := gl.CreateShader(kind)
	cSource, cSourceFreeFn := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, cSource, nil)
	cSourceFreeFn()

	// Compile the shader
	gl.CompileShader(shader)

	// Check for a compilation error
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		// Get the length of the error message
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		// Retrieve the error message
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf(log)
	}

	return shader, nil
}

// LinkProgram links together a shader program, checking for any errors.
func linkProgram(program uint32) error {
	// Link the shader program
	gl.LinkProgram(program)

	// Check if there was a link error
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		// Get the length of the error message
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		// Retrieve the error message
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return fmt.Errorf(log)
	}

	return nil
}

// LoadTexture reads texture data from memory and uploads it to a GPU texture
// for use with OpenGL.
func LoadTexture(img *image.RGBA, slot uint32) uint32 {
	// Generate the texture
	var texture uint32
	gl.GenTextures(1, &texture)

	// Bind the texture to the desired slot
	gl.ActiveTexture(gl.TEXTURE0 + slot)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Upload the texture data
	width := int32(img.Bounds().Max.X - img.Bounds().Min.X)
	height := int32(img.Bounds().Max.Y - img.Bounds().Min.Y)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(&img.Pix[0]))

	// Disable wrapping
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)

	// Disable antialiasing
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	return texture
}

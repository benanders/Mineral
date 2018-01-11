package render

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// LoadShaders compiles a vertex and fragment shader from a string, creates a
// new OpenGL shader program, attaches the two shaders, and links the program.
func LoadShaders(vertexSource, fragmentSource string) (uint32, error) {
	// Compile the vertex and fragment shaders
	vertex, err := compileShader(gl.VERTEX_SHADER, vertexSource)
	if err != nil {
		return 0, err
	}
	fragment, err := compileShader(gl.FRAGMENT_SHADER, fragmentSource)
	if err != nil {
		return 0, err
	}

	// Create the program and attach the vertex and fragment shaders
	program := gl.CreateProgram()
	gl.AttachShader(program, vertex)
	gl.AttachShader(program, fragment)

	// Link the program
	err = linkProgram(program)
	if err != nil {
		return 0, err
	}

	return program, nil
}

// LoadShader compiles a shader from a string, checking for any compilation
// errors.
func compileShader(kind uint32, source string) (uint32, error) {
	// Create the shader and attach the source code to it
	shader := gl.CreateShader(kind)
	cSource, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, cSource, nil)
	free()

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
		return 0, fmt.Errorf("failed to compile shader: %v", log)
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
		return fmt.Errorf("failed to link shader: %v", log)
	}

	return nil
}

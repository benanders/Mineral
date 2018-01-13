package util

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

//
//  OpenGL Utilities
//

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

//
//  Math Utilities
//

// AABB is an axis aligned bounding box, used for all collision detection.
type AABB struct {
	Center mgl32.Vec3
	Size   mgl32.Vec3
}

// MinX returns the minimum x bound for the AABB.
func (a AABB) MinX() float32 { return a.Center.X() - a.Size.X()/2.0 }

// MaxX returns the maximum x bound for the AABB.
func (a AABB) MaxX() float32 { return a.Center.X() + a.Size.X()/2.0 }

// MinY returns the minimum y bound for the AABB.
func (a AABB) MinY() float32 { return a.Center.Y() - a.Size.Y()/2.0 }

// MaxY returns the maximum y bound for the AABB.
func (a AABB) MaxY() float32 { return a.Center.Y() + a.Size.Y()/2.0 }

// MinZ returns the minimum z bound for the AABB.
func (a AABB) MinZ() float32 { return a.Center.Z() - a.Size.Z()/2.0 }

// MaxZ returns the maximum z bound for the AABB.
func (a AABB) MaxZ() float32 { return a.Center.Z() + a.Size.Z()/2.0 }

// Offset moves the position of the AABB by the given delta.
func (a *AABB) Offset(delta mgl32.Vec3) {
	a.Center = a.Center.Add(delta)
}

// Intersects returns true if the two AABBs overlap.
func (a AABB) Intersects(b AABB) bool {
	return a.MinX() < b.MaxX() && a.MaxX() > b.MinX() &&
		a.MinY() < b.MaxY() && a.MaxY() > b.MinY() &&
		a.MinZ() < b.MaxZ() && a.MaxZ() > b.MinZ()
}

// IntersectionX returns the overlap between two AABBs along the X axis.
func (a AABB) IntersectionX(b AABB) float32 {
	// Due to floating point precision error, we need to increase the overlap
	// slightly in order to resolve collisions properly
	if a.MaxX()-b.MinX() < b.MaxX()-a.MinX() {
		return math.Nextafter32(a.MaxX()-b.MinX(), float32(math.Inf(1)))
	}
	return math.Nextafter32(a.MinX()-b.MaxX(), float32(math.Inf(-1)))
}

// IntersectionY returns the overlap between two AABBs along the Y axis.
func (a AABB) IntersectionY(b AABB) float32 {
	if a.MaxY()-b.MinY() < b.MaxY()-a.MinY() {
		return math.Nextafter32(a.MaxY()-b.MinY(), float32(math.Inf(1)))
	}
	return math.Nextafter32(a.MinY()-b.MaxY(), float32(math.Inf(-1)))
}

// IntersectionZ returns the overlap between two AABBs along the Z axis.
func (a AABB) IntersectionZ(b AABB) float32 {
	if a.MaxZ()-b.MinZ() < b.MaxZ()-a.MinZ() {
		return math.Nextafter32(a.MaxZ()-b.MinZ(), float32(math.Inf(1)))
	}
	return math.Nextafter32(a.MinZ()-b.MaxZ(), float32(math.Inf(-1)))
}

// Lerp performs linear interpolation between the starting and ending values,
// based on the given amount.
func Lerp(start, end, amount float32) float32 {
	return start*(1.0-amount) + end*amount
}

// Clamp restricts a value between a minimum and maximum value.
func Clamp(value, min, max float32) float32 {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

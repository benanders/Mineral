package sky

import (
	"log"
	"math"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/benanders/mineral/camera"
	"github.com/benanders/mineral/util"
	"github.com/benanders/mineral/world"
)

// The temperature throughout the world (influences the sky, fog, and sunrise
// colors slightly). Individual biomes will have different temperatures in the
// future.
const worldTemperature = 0.5

// Sky is responsible for drawing the background sky in the game.
type Sky struct {
	skyPlane     skyPlane
	sunrisePlane sunrisePlane
}

// RenderInfo stores a bunch of information required by the sky renderer in
// order to draw the sky.
type RenderInfo struct {
	WorldTime    float64
	Camera       *camera.Camera
	RenderRadius uint
	LookDir      mgl32.Vec3
}

// SkyPlane stores information about the blue ceiling plane and the dark blue
// void plane present in the sky.
type skyPlane struct {
	skyVao, skyVbo   uint32
	voidVao, voidVbo uint32
	program          uint32
	mvpUnf           int32
	colorUnf         int32
	fogColorUnf      int32
	farPlaneUnf      int32
}

// SunrisePlane stores information about the red/orange sunrise/sunset plane
// present in the sky during sunrise and sunset.
type sunrisePlane struct {
	vao, vbo        uint32
	program         uint32
	mvpUnf          int32
	sunriseColorUnf int32
}

const (
	// SkyVertexShader stores the source code for the vertex shader for the
	// sky plane.
	skyVertexShader = `
#version 330

uniform mat4 mvp;

in vec3 position;
out vec3 fragPos;

void main() {
	gl_Position = mvp * vec4(position, 1.0);
	fragPos = position;
}
`

	// SkyFragmentShader stores the source code for the fragment shader for the
	// sky plane.
	skyFragmentShader = `
#version 330

uniform vec3 skyColor;
uniform vec3 fogColor;
uniform float farPlane;

in vec3 fragPos;
out vec4 color;

void main() {
	// Use the position of the fragment to calculate the fog strength
	float fog_strength = length(fragPos) / (farPlane * 0.8);
	fog_strength = clamp(fog_strength, 0.0, 1.0);

	// Modulate between the sky and fog colors by the fog strength factor
	color = vec4(mix(skyColor, fogColor, fog_strength), 1.0);
}
`

	// SunriseVertexShader stores the source code for the vertex shader for the
	// sunrise plane.
	sunriseVertexShader = `
#version 330

uniform mat4 mvp;
uniform vec4 sunriseColor;

in vec3 position;
in float alpha;
out float fragAlpha;

void main() {
	// Modulate the z component of the position by the alpha component of the
	// sunrise color
	gl_Position = mvp * vec4(position.xy, position.z * sunriseColor.a, 1.0);
	fragAlpha = alpha;
}
`

	// SunriseFragmentShader stores the source code for the fragment shader for
	// the sunrise plane.
	sunriseFragmentShader = `
#version 330

uniform vec4 sunriseColor;

in float fragAlpha;
out vec4 color;

void main() {
	// The alpha multiplier is 1 for the centre point of the fan, and 0 for all
	// the points on the rim. This means the sunrise color fades to nothing on
	// the rim, and has an alpha component of sunriseColor.a at the centre
	color = vec4(sunriseColor.rgb, sunriseColor.a * fragAlpha);
}
`
)

// NewSky creates a new sky renderer instance.
func NewSky() *Sky {
	return &Sky{newSkyPlane(), newSunrisePlane()}
}

// Destroy releases all the resources allocated by the sky renderer.
func (s *Sky) Destroy() {
	s.skyPlane.destroy()
	s.sunrisePlane.destroy()
}

// NewSkyPlane builds the vertex data and allocates the required OpenGL
// resources for the sky plane.
func newSkyPlane() skyPlane {
	// Create the shader progarm
	program, err := util.LoadShaders(skyVertexShader, skyFragmentShader)
	if err != nil {
		log.Fatalln(err)
	}
	gl.UseProgram(program)

	// Cache the locations of uniforms
	mvpUnf := gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	colorUnf := gl.GetUniformLocation(program, gl.Str("skyColor\x00"))
	fogColorUnf := gl.GetUniformLocation(program, gl.Str("fogColor\x00"))
	farPlaneUnf := gl.GetUniformLocation(program, gl.Str("farPlane\x00"))

	// Vertex data for the sky plane
	skyVertices := [...]float32{
		-384.0, 16.0, -384.0, // The size of the sky plane must be larger
		384.0, 16.0, -384.0, // than the far plane distance, or else the
		-384.0, 16.0, 384.0, // sky will look noticeably square.
		384.0, 16.0, 384.0,
	}

	// Vertex data for the void plane
	voidVertices := [...]float32{
		-384.0, -16.0, -384.0, // Swap the winding order from the sky plane so
		-384.0, -16.0, 384.0, // GL_CULL_FACE still works.
		384.0, -16.0, -384.0,
		384.0, -16.0, 384.0,
	}

	// Create the sky and void planes
	skyVao, skyVbo := genPlane(program, skyVertices[:])
	voidVao, voidVbo := genPlane(program, voidVertices[:])

	// Create the sky plane object holding it all together
	return skyPlane{skyVao, skyVbo, voidVao, voidVbo, program, mvpUnf,
		colorUnf, fogColorUnf, farPlaneUnf}
}

// Generates the sky or void plane VAO and VBO, and enables the vertex
// attributes.
func genPlane(program uint32, vertices []float32) (vao, vbo uint32) {
	// Create the VAO
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Create the VBO and populate it with data
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER,
		int(unsafe.Sizeof(float32(0.0)))*len(vertices),
		gl.Ptr(vertices[:]),
		gl.STATIC_DRAW)

	// Enable the position attribute
	posAttr := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	gl.EnableVertexAttribArray(posAttr)
	gl.VertexAttribPointer(posAttr, 3, gl.FLOAT, false, 0, nil)
	return
}

// Destroy releases all the resources allocated by the sky plane.
func (p *skyPlane) destroy() {
	gl.DeleteProgram(p.program)
	gl.DeleteVertexArrays(1, &p.skyVao)
	gl.DeleteVertexArrays(1, &p.voidVao)
	gl.DeleteBuffers(1, &p.skyVbo)
	gl.DeleteBuffers(1, &p.voidVbo)
}

// NewSunrisePlane builds the vertex data and allocates the required OpenGL
// resources for the sunrise plane.
func newSunrisePlane() sunrisePlane {
	// Create the VAO
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Create the VBO and populate it with data
	vertices := genSunrisePlaneVertices()
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(float32(0.0))*18*4),
		gl.Ptr(vertices[:]), gl.STATIC_DRAW)

	// Create the shader progarm
	program, err := util.LoadShaders(sunriseVertexShader, sunriseFragmentShader)
	if err != nil {
		log.Fatalln(err)
	}
	gl.UseProgram(program)

	// Cache the locations of uniforms
	mvpUnf := gl.GetUniformLocation(program, gl.Str("mvp\x00"))
	colorUnf := gl.GetUniformLocation(program, gl.Str("sunriseColor\x00"))

	// Enable the position attribute
	posAttr := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	gl.EnableVertexAttribArray(posAttr)
	gl.VertexAttribPointer(posAttr, 3, gl.FLOAT, false,
		int32(4*unsafe.Sizeof(float32(0.0))), nil)

	// Enable the alpha multiplier attribute
	alphaAttr := uint32(gl.GetAttribLocation(program, gl.Str("alpha\x00")))
	offset := 3 * unsafe.Sizeof(float32(0.0))
	gl.EnableVertexAttribArray(alphaAttr)
	gl.VertexAttribPointer(alphaAttr, 1, gl.FLOAT, false,
		int32(4*unsafe.Sizeof(float32(0.0))), // stride of 4
		gl.PtrOffset(int(offset)))            // offset of 3

	return sunrisePlane{vao, vbo, program, mvpUnf, colorUnf}
}

// GenSunrisePlaneVertices builds the vertex data array for the sunrise plane.
func genSunrisePlaneVertices() []float32 {
	// Generate the vertex data
	var vertices [18 * 4]float32
	vertices[0] = 0.0   // position.x
	vertices[1] = 100.0 // position.y
	vertices[2] = 0.0   // position.z
	vertices[3] = 1.0   // alpha multiplier

	for i := 0; i <= 16; i++ {
		// The original minecraft source modulates the z component by the alpha
		// of the current sunrise/sunset color. Since the alpha changes every
		// frame, we do this in the vertex shader to reduce runtime overhead.
		angle := float64(i) * 2.0 * math.Pi / 16.0
		sin, cos := math.Sincos(angle)
		vertices[(i+1)*4] = float32(sin) * 120.0   // position.x
		vertices[(i+1)*4+1] = float32(cos) * 120.0 // position.y
		vertices[(i+1)*4+2] = -float32(cos) * 40.0 // position.z
		vertices[(i+1)*4+3] = 0.0                  // alpha multiplier
	}

	return vertices[:]
}

// Destroy releases all the resources allocated by the sunrise plane.
func (p *sunrisePlane) destroy() {
	gl.DeleteProgram(p.program)
	gl.DeleteVertexArrays(1, &p.vao)
	gl.DeleteBuffers(1, &p.vbo)
}

// Color represents a color as red, green, and blue color components.
type color struct {
	r, g, b float64
}

// Returns the color components as float32s, suitable for use with OpenGL.
func (c color) R() float32 { return float32(c.r) }
func (c color) G() float32 { return float32(c.g) }
func (c color) B() float32 { return float32(c.b) }

// HsvToRgb converts a color from HSV color space to RGB color space.
func hsvToRgb(h, s, v float64) color {
	option := int(h*6.0) % 6
	factor := h*6.0 - float64(option)
	a := v * (1.0 - s)
	b := v * (1.0 - factor*s)
	c := v * (1.0 - (1.0-factor)*s)
	switch option {
	case 0:
		return color{v, c, a}
	case 1:
		return color{b, v, a}
	case 2:
		return color{a, v, c}
	case 3:
		return color{a, b, v}
	case 4:
		return color{c, a, v}
	case 5:
		return color{v, a, b}
	}
	return color{}
}

// The celestial angle is proportional to the angle that the sun makes with the
// horizon. It is a value between 0 and 1 representing the time of day.
func getCelestialAngle(worldTime float64) float64 {
	// Since world time is measured in days, the progress through the current
	// day is just the fractional part of `worldTime`
	dayProgress := worldTime - float64(uint64(worldTime))

	// We subtract 0.25 so that the start of the day (worldTime = 0) represents
	// sunrise, rather than midnight
	dayProgress -= 0.25

	// Wrap the day progress to some value between 0 and 1
	if dayProgress < 0.0 {
		dayProgress += 1.0
	} else if dayProgress > 1.0 {
		dayProgress -= 1.0
	}

	// This is the magical celestial angle calculation from the Minecraft source
	celestialAngle := 0.5 * (1.0 - math.Cos(dayProgress*math.Pi))
	return dayProgress + (celestialAngle-dayProgress)/3.0
}

// The sky color is used for the sky plane, and is normally a slightly darker
// blue than the fog color.
func getSkyColor(celestialAngle float64) color {
	// Calculate the base color based on the temperature
	temperature := util.Clamp64(worldTemperature/3.0, -1.0, 1.0)
	base := hsvToRgb(
		0.62222224-temperature*0.05,
		0.5+temperature*0.1,
		1.0)

	// Calculate the brightness multiplier
	brightness := math.Cos(celestialAngle*math.Pi*2.0)*2.0 + 0.5
	brightness = util.Clamp64(brightness, 0.0, 1.0)

	// Calculate the final color
	return color{base.r * brightness, base.g * brightness,
		base.b * brightness}
}

// The void color is used for the void plane, and is normally a deeper blue
// than the sky color.
func getVoidColor(celestialAngle float64) color {
	// Calculate the void plane color based off the sky color
	skyColor := getSkyColor(celestialAngle)
	return color{
		skyColor.r*0.2 + 0.04,
		skyColor.g*0.2 + 0.04,
		skyColor.b*0.6 + 0.1}
}

// Calculates the color of the sunrise/sunset, based on the current celestial
// angle.
func getSunriseColor(celestialAngle float64) (color, float64) {
	// Calculate time of day multiplier
	multiplier := math.Cos(celestialAngle * 2.0 * math.Pi)

	// Only apply the sunrise/sunset color if the time of day is right
	if multiplier >= -0.4 && multiplier <= 0.4 {
		phase := multiplier*1.25 + 0.5
		sqrtAlpha := math.Sin(phase*math.Pi)*0.99 + 0.01
		sunriseColor := color{
			phase*0.3 + 0.7,
			phase*phase*0.7 + 0.2,
			0.2}
		return sunriseColor, sqrtAlpha * sqrtAlpha
	}
	return color{}, 0.0
}

// Calculates the background fog color, including the influence of looking
// towards the sunrise/sunset.
func getFogColor(celestialAngle float64, renderRadius uint,
	lookDir mgl32.Vec3) color {
	// Calculate the brightness multiplier
	brightness := math.Cos(celestialAngle*math.Pi*2.0)*2.0 + 0.5
	brightness = util.Clamp64(brightness, 0.0, 1.0)

	// Calculate the fog color using some magic numbers
	fogColor := color{
		0.7529412 * (brightness*0.94 + 0.06),
		0.84705883 * (brightness*0.94 + 0.06),
		1.0 * (brightness*0.91 + 0.09)}

	// Modify the fog with the sunrise/sunset color
	if renderRadius >= 4 {
		// Get a vector whose direction depends on whether this is a sunrise or
		// sunset
		sinAngle := math.Sin(celestialAngle * math.Pi * 2.0)
		var sunDir mgl32.Vec3
		if sinAngle < 0.0 {
			sunDir = mgl32.Vec3{-1.0, 0.0, 0.0}
		} else {
			sunDir = mgl32.Vec3{1.0, 0.0, 0.0}
		}

		// Calculate the look direction multiplier (player facing more towards
		// the sunrise/sunset makes it look more intense)
		lookMultiplier := math.Max(float64(lookDir.Dot(sunDir)), 0.0)

		// Get the sunrise/sunset color
		sunriseColor, alpha := getSunriseColor(celestialAngle)

		// Modify the fog color based on the sunrise/sunset color
		lookMultiplier *= alpha
		fogColor.r = util.Lerp64(fogColor.r, sunriseColor.r, lookMultiplier)
		fogColor.g = util.Lerp64(fogColor.g, sunriseColor.g, lookMultiplier)
		fogColor.b = util.Lerp64(fogColor.b, sunriseColor.b, lookMultiplier)
	}

	// Modify the fog color with the sky color based on the render radius
	sky := getSkyColor(celestialAngle)
	fractionalRadius := float64(renderRadius) / float64(world.MaxRenderRadius)
	sightFactor := 1.0 - math.Pow(fractionalRadius*0.75+0.25, 0.25)
	fogColor.r += (sky.r - fogColor.r) * sightFactor
	fogColor.g += (sky.g - fogColor.g) * sightFactor
	fogColor.b += (sky.b - fogColor.b) * sightFactor
	return fogColor
}

// Clears the screen to the current fog color.
func (s *Sky) renderBackground(info RenderInfo) {
	// Get the current fog color
	celestialAngle := getCelestialAngle(info.WorldTime)
	fogColor := getFogColor(celestialAngle, info.RenderRadius, info.LookDir)

	// Clear the screen
	gl.ClearColor(fogColor.R(), fogColor.G(), fogColor.B(), 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// Renders the sky plane based on the camera's orientation matrix, and the
// current sky and fog colors.
func (p *skyPlane) renderSky(info RenderInfo) {
	// Set the current shader program to the sky plane program
	gl.UseProgram(p.program)

	// Set the shader's MVP uniform to the camera's orientation matrix
	gl.UniformMatrix4fv(p.mvpUnf, 1, false, &info.Camera.Orientation[0])

	// Set the color of the sky plane to the sky color
	celestialAngle := getCelestialAngle(info.WorldTime)
	skyColor := getSkyColor(celestialAngle)
	gl.Uniform3f(p.colorUnf, skyColor.R(), skyColor.G(), skyColor.B())

	// Set the fog color uniform
	fogColor := getFogColor(celestialAngle, info.RenderRadius, info.LookDir)
	gl.Uniform3f(p.fogColorUnf, fogColor.R(), fogColor.G(), fogColor.B())

	// Set the far plane distance, used for fog calculations
	gl.Uniform1f(p.farPlaneUnf, info.Camera.FarPlane)

	// Render the sky plane
	gl.BindVertexArray(p.skyVao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

// Renders the void plane based on the camera's orientation matrix, and the
// current void and fog colors.
func (p *skyPlane) renderVoid(info RenderInfo) {
	// Only change the sky color uniform from rendering the sky plane above,
	// to the void color
	celestialAngle := getCelestialAngle(info.WorldTime)
	voidColor := getVoidColor(celestialAngle)
	gl.Uniform3f(p.colorUnf, voidColor.R(), voidColor.G(), voidColor.B())

	// Render the sky plane
	gl.BindVertexArray(p.voidVao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

// Renders the sunrise/sunset plane based on the camera's orientation matrix,
// and the current sunrise/sunset color.
func (p *sunrisePlane) render(info RenderInfo) {
	// Set the current shader program to the sunrise plane program
	gl.UseProgram(p.program)

	// Calculate a rotation matrix based on whether it's currently sunrise or
	// sunset (to change where the sunrise plane appears in the sky)
	celestialAngle := getCelestialAngle(info.WorldTime)
	var todAngle float32 // tod stands for time of day
	if math.Sin(celestialAngle*math.Pi*2.0) < 0.0 {
		todAngle = math.Pi
	} else {
		todAngle = 0.0
	}
	todRot := mgl32.HomogRotate3D(todAngle, mgl32.Vec3{0.0, 0.0, 1.0})

	// Set the shader's MVP uniform to the camera's orientation matrix
	xRot := mgl32.HomogRotate3D(math.Pi/2.0, mgl32.Vec3{1.0, 0.0, 0.0})
	zRot := mgl32.HomogRotate3D(math.Pi/2.0, mgl32.Vec3{0.0, 0.0, 1.0})
	mvp := info.Camera.Orientation.Mul4(xRot.Mul4(todRot.Mul4(zRot)))
	gl.UniformMatrix4fv(p.mvpUnf, 1, false, &mvp[0])

	// Set the sunrise color uniform
	color, alpha := getSunriseColor(celestialAngle)
	gl.Uniform4f(p.sunriseColorUnf, color.R(), color.G(), color.B(),
		float32(alpha))

	// Render the sunrise plane with linear alpha blending enabled
	gl.Enable(gl.BLEND)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ZERO)

	// Render the sunrise plane
	gl.BindVertexArray(p.vao)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 18)

	// Reset the OpenGL state
	gl.Disable(gl.BLEND)
}

// Render clears the color buffer to the fog color, renders the sky plane,
// sunrise/sunset plane, sun/moon, stars, and void plane.
func (s *Sky) Render(info RenderInfo) {
	// Enable some OpenGL configuration
	gl.Enable(gl.CULL_FACE)

	// Render components of the sky separately
	s.renderBackground(info)
	s.skyPlane.renderSky(info)
	s.skyPlane.renderVoid(info)
	s.sunrisePlane.render(info)

	// Reset the OpenGL configuration
	gl.Disable(gl.CULL_FACE)
}

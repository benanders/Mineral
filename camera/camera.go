package camera

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// ViewPoint tells the camera enough information about an object, such that the
// scene can be viewed from that object's perspective.
type ViewPoint interface {
	Sight() mgl32.Vec3
	EyePosition() mgl32.Vec3
}

// DefaultFov is the default field of view for the camera, set to 60 degrees
// converted to radians.
const DefaultFov = 60.0 * float32(math.Pi) / 180.0 // 60 degrees in radians

// Camera keeps track of the model, view, projection, and orientation matrices,
// which define the perspective from which the scene is viewed.
type Camera struct {
	FarPlane    float32
	Projection  mgl32.Mat4
	View        mgl32.Mat4
	Orientation mgl32.Mat4
}

// Perspective sets up the camera's perspective projection with the given
// parameters. `fov` is in radians.
func (c *Camera) Perspective(fov, aspect, near, far float32) {
	c.FarPlane = far
	c.Projection = mgl32.Perspective(fov, aspect, near, far)
}

// Follow updates the camera's view and orientation matrices so that the scene
// is now viewed from the perspective of the given entity.
func (c *Camera) Follow(viewPoint ViewPoint) {
	eye := viewPoint.EyePosition()
	sight := viewPoint.Sight()
	up := mgl32.Vec3{0.0, 1.0, 0.0}

	// Orientation matrix (no translation, just rotation)
	orientation := mgl32.LookAtV(sight, mgl32.Vec3{}, up)
	c.Orientation = c.Projection.Mul4(orientation)

	// View matrix (incorporates position)
	view := mgl32.LookAtV(eye, eye.Sub(sight), up)
	c.View = c.Projection.Mul4(view)
}

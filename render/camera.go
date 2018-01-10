package render

import (
	"github.com/benanders/mineral/entity"
	"github.com/go-gl/mathgl/mgl32"
)

// Camera keeps track of the model, view, projection, and orientation matrices,
// which define the perspective from which the scene is viewed.
type Camera struct {
	farPlane    float32
	projection  mgl32.Mat4
	view        mgl32.Mat4
	orientation mgl32.Mat4
}

// Perspective sets up the camera's perspective projection with the given
// parameters.
func (c *Camera) Perspective(fov, aspect, near, far float32) {
	c.farPlane = far
	c.projection = mgl32.Perspective(fov, aspect, near, far)
}

// Follow updates the camera's view and orientation matrices so that the scene
// is now viewed from the perspective of the given entity.
func (c *Camera) Follow(entity *entity.Entity) {
	orientation := mgl32.LookAtV(entity.Sight(), mgl32.Vec3{0.0, 0.0, 0.0},
		mgl32.Vec3{0.0, 1.0, 0.0})
	c.orientation = c.projection.Mul4(orientation)
	view := mgl32.LookAtV(entity.Sight(), entity.Position(),
		mgl32.Vec3{0.0, 1.0, 0.0})
	c.view = c.projection.Mul4(view)
}

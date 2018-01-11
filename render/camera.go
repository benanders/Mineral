package render

import (
	"github.com/benanders/mineral/entity"

	"github.com/go-gl/mathgl/mgl32"
)

// Camera keeps track of the model, view, projection, and orientation matrices,
// which define the perspective from which the scene is viewed.
type Camera struct {
	FarPlane    float32
	Projection  mgl32.Mat4
	View        mgl32.Mat4
	Orientation mgl32.Mat4
}

// Perspective sets up the camera's perspective projection with the given
// parameters.
func (c *Camera) Perspective(fov, aspect, near, far float32) {
	c.FarPlane = far
	c.Projection = mgl32.Perspective(fov, aspect, near, far)
}

// Follow updates the camera's view and orientation matrices so that the scene
// is now viewed from the perspective of the given entity.
func (c *Camera) Follow(entity *entity.Entity) {
	pos := entity.Position()
	sight := entity.Sight()

	// Orientation matrix (no translation, just rotation)
	orientation := mgl32.LookAtV(sight, mgl32.Vec3{0.0, 0.0, 0.0},
		mgl32.Vec3{0.0, 1.0, 0.0})
	c.Orientation = c.Projection.Mul4(orientation)

	// View matrix (incorporates position)
	view := mgl32.LookAtV(pos, pos.Sub(sight), mgl32.Vec3{0.0, 1.0, 0.0})
	c.View = c.Projection.Mul4(view)
}

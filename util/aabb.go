package util

import (
	"github.com/go-gl/mathgl/mgl32"
)

// AABB is an axis aligned bounding box, used for all collision detection.
type AABB struct {
	Center mgl32.Vec3
	Size   mgl32.Vec3
}

func (a AABB) MinX() float32 { return a.Center.X() - a.Size.X()/2.0 }
func (a AABB) MaxX() float32 { return a.Center.X() + a.Size.X()/2.0 }

func (a AABB) MinY() float32 { return a.Center.Y() - a.Size.Y()/2.0 }
func (a AABB) MaxY() float32 { return a.Center.Y() + a.Size.Y()/2.0 }

func (a AABB) MinZ() float32 { return a.Center.Z() - a.Size.Z()/2.0 }
func (a AABB) MaxZ() float32 { return a.Center.Z() + a.Size.Z()/2.0 }

// Offset moves the position of the AABB by the given delta.
func (a *AABB) Offset(delta mgl32.Vec3) { a.Center = a.Center.Add(delta) }

// IsOverlapping returns true if the two AABBs overlap.
func (a AABB) IsOverlapping(b AABB) bool {
	return a.MinX() < b.MaxX() && a.MaxX() > b.MinX() &&
		a.MinY() < b.MaxY() && a.MaxY() > b.MinY() &&
		a.MinZ() < b.MaxZ() && a.MaxZ() > b.MinZ()
}

// Overlap returns the overlap between two colliding AABBs along each axis.
func (a AABB) Overlap(b AABB) mgl32.Vec3 {
	var x, y, z float32

	// X axis overlap
	if a.MaxX()-b.MinX() < b.MaxX()-a.MinX() {
		x = a.MaxX() - b.MinX()
	} else {
		x = a.MinX() - b.MaxX()
	}

	// Y axis overlap
	if a.MaxY()-b.MinY() < b.MaxY()-a.MinY() {
		y = a.MaxY() - b.MinY()
	} else {
		y = a.MinY() - b.MaxY()
	}

	// Z axis overlap
	if a.MaxZ()-b.MinZ() < b.MaxZ()-a.MinZ() {
		z = a.MaxZ() - b.MinZ()
		println("positive", z)
	} else {
		z = a.MinZ() - b.MaxZ()
		println("negative", z)
	}

	return mgl32.Vec3{x, y, z}
}

package math

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

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

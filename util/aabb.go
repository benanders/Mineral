package util

import "github.com/go-gl/mathgl/mgl32"

// AABB is an axis aligned bounding box, used for all collision detection.
type AABB struct {
	Center mgl32.Vec3
	Size   mgl32.Vec3
}

func (a *AABB) MinX() float32 { return a.Center.X() - a.Size.X()/2.0 }
func (a *AABB) MaxX() float32 { return a.Center.X() + a.Size.X()/2.0 }

func (a *AABB) MinY() float32 { return a.Center.Y() - a.Size.Y()/2.0 }
func (a *AABB) MaxY() float32 { return a.Center.Y() + a.Size.Y()/2.0 }

func (a *AABB) MinZ() float32 { return a.Center.Z() - a.Size.Z()/2.0 }
func (a *AABB) MaxZ() float32 { return a.Center.Z() + a.Size.Z()/2.0 }

// Offset moves the position of the AABB by the given delta.
func (a *AABB) Offset(delta mgl32.Vec3) { a.Center = a.Center.Add(delta) }

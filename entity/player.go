package entity

import (
	"github.com/benanders/mineral/math"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	// PlayerMoveSpeed is the default speed at which the player can move.
	playerMoveSpeed = 0.1

	// PlayerLookSpeed is the default speed at which the player can look
	// around.
	playerLookSpeed = 0.003
)

// Player is an entity controlled by the user, which the camera follows as they
// move.
type Player struct {
	Entity
}

// NewPlayer creates a new instance of the player with an initial position and
// rotation.
func NewPlayer(center mgl32.Vec3, rotation mgl32.Vec2) *Player {
	// Default player size is 0.6 x 1.8 x 0.6 blocks
	aabb := math.AABB{Center: center, Size: mgl32.Vec3{0.6, 1.8, 0.6}}
	entity := NewEntity(aabb, rotation, playerMoveSpeed, playerLookSpeed)
	p := Player{*entity}
	p.updateAxes()
	return &p
}

// Sight implements the camera.ViewPoint interface for the player.
func (p *Player) Sight() mgl32.Vec3 {
	return p.Entity.Sight
}

// EyePosition implements the camera.ViewPoint interface for the player.
func (p *Player) EyePosition() mgl32.Vec3 {
	// The player's eye sits slightly below the top of their AABB, 90% of the
	// way up their body
	return mgl32.Vec3{p.AABB.Center.X(),
		p.AABB.Center.Y() + p.AABB.Size.Y()*0.4,
		p.AABB.Center.Z()}
}

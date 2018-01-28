package game

import (
	"time"

	"github.com/benanders/mineral/camera"
	"github.com/benanders/mineral/entity"
	"github.com/benanders/mineral/sky"
	"github.com/benanders/mineral/world"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

// Game stores all the required state information while the game is running.
type Game struct {
	window *sdl.Window

	sky   *sky.Sky
	world *world.World

	camera           *camera.Camera
	player           *entity.Player
	playerController entity.Controller

	startTime time.Time
}

// New creates a new game state.
func New(window *sdl.Window) *Game {
	g := Game{window: window, startTime: time.Now()}

	g.sky = sky.New()
	g.world = world.New(16)
	g.world.GenChunk(0, 0)

	g.player = entity.NewPlayer(mgl32.Vec3{0.0, 5.0, 0.0}, mgl32.Vec2{})
	g.playerController = entity.NewInputController()

	w, h := sdl.GLGetDrawableSize(window)
	aspect := float32(w) / float32(h)
	g.camera = &camera.Camera{}
	g.camera.Perspective(camera.Fov, aspect, camera.Near, camera.Far)
	g.camera.Follow(g.player)

	return &g
}

// Destroy frees all resources allocated by the game state.
func (g *Game) Destroy() {
	g.world.Destroy()
	g.sky.Destroy()
}

// HandleEvent processes a user input event.
func (g *Game) HandleEvent(evt sdl.Event) {
	g.playerController.HandleEvent(evt)
}

// Update advances the game state. It's called at a fixed time step, in order
// to simplify some of the mechanics of the code (particularly the physics).
func (g *Game) Update() {
	// Checks for completed chunk load requests
	g.world.Update()

	// Update the player's movement
	g.player.ApplyMovementAndResolveCollisions(g.world)

	// Get the camera to follow the player
	g.playerController.Update(g.player)
	g.camera.Follow(g.player)
}

// Render draws the game to the screen. It's called as fast as possible. Render
// frames are dropped (slowing the visible FPS) if updating the game takes
// longer than the alloted time.
func (g *Game) Render() {
	// Sky is rendered first, underneath everything else
	g.sky.Render(sky.RenderInfo{
		WorldTime:    0.0,
		Camera:       g.camera,
		RenderRadius: g.world.RenderRadius,
		LookDir:      g.player.Sight()})

	// The world is rendered on top of the sky
	g.world.Render(world.RenderInfo{
		Camera: g.camera,
	})
}

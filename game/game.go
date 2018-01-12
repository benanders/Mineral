package game

import (
	"time"

	"github.com/benanders/mineral/camera"
	"github.com/benanders/mineral/entity"
	"github.com/benanders/mineral/entity/ctrl"
	"github.com/benanders/mineral/sky"
	"github.com/benanders/mineral/world"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

// Game stores all the required state information while the game is running.
type Game struct {
	window     *sdl.Window
	player     *entity.Player
	playerCtrl ctrl.Controller
	camera     *camera.Camera
	sky        *sky.Sky
	world      *world.World

	startTime time.Time
	worldTime float64
}

// New creates a new game state.
func New(window *sdl.Window) *Game {
	g := Game{window: window, startTime: time.Now()}

	// World
	g.world = world.New(16)
	g.world.LoadChunk(0, 0)

	// Sky
	g.sky = sky.NewSky()

	// Player, and the player's input controller
	g.player = entity.NewPlayer(mgl32.Vec3{0.0, 5.0, 0.0}, mgl32.Vec2{})
	g.playerCtrl = ctrl.NewInputCtrl()

	// Camera
	w, h := sdl.GLGetDrawableSize(window)
	aspect := float32(w) / float32(h)
	g.camera = &camera.Camera{}
	g.camera.Perspective(camera.DefaultFov, aspect, 0.1, 256.0)
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
	// Pass the event onto the player's input controller
	g.playerCtrl.HandleEvent(evt)
}

// Update advances the game state. It's called at a fixed time step, in order
// to simplify some of the mechanics of the code.
func (g *Game) Update() {
	// Update the world
	g.world.Update()

	// Resolve collisions between the player and the world
	g.player.ApplyMovementAndResolveCollisions(g.world)

	// Update the camera, making it follow the position and look direction of
	// the player
	g.playerCtrl.Update(g.player)
	g.camera.Follow(g.player)

	// For debugging only
	if g.playerCtrl.(*ctrl.InputCtrl).IsKeyDown[sdl.SCANCODE_UP] {
		g.worldTime += 0.005
	} else if g.playerCtrl.(*ctrl.InputCtrl).IsKeyDown[sdl.SCANCODE_DOWN] {
		g.worldTime -= 0.005
	}
}

// Render draws the game to the screen. It's called at a variable time step
// (basically as fast as possible).
func (g *Game) Render() {
	// Render the sky
	g.sky.Render(sky.RenderInfo{
		WorldTime:    g.worldTime,
		Camera:       g.camera,
		RenderRadius: g.world.RenderRadius,
		LookDir:      g.player.Sight()})

	// Render the world
	g.world.Render(world.RenderInfo{Camera: g.camera})
}

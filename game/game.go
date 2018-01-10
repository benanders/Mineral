package game

import (
	"math"
	"time"

	"github.com/benanders/mineral/entity"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/benanders/mineral/render"
	"github.com/veandco/go-sdl2/sdl"
)

// Game stores all the required state information while the game is running.
type Game struct {
	window     *sdl.Window
	startTime  time.Time
	player     *entity.Entity
	playerCtrl *entity.InputCtrl
	camera     *render.Camera
	sky        *render.Sky

	worldTime float64
}

// New creates a new game state.
func New(window *sdl.Window) *Game {
	g := Game{window: window, startTime: time.Now(), camera: &render.Camera{}}

	// Set up the player and its input controller
	g.player = entity.NewEntity(mgl32.Vec3{0.0, 5.0, 0.0}, mgl32.Vec2{0.0, 0.0})
	g.playerCtrl = entity.NewInputCtrl(0.5, 0.003)

	// Create the camera and make it follow the player
	fov := 70.0 * float32(math.Pi) / 180.0 // 70 degrees in radians
	w, h := sdl.GLGetDrawableSize(window)
	aspect := float32(w) / float32(h)
	g.camera.Perspective(fov, aspect, 0.1, 256.0)
	g.camera.Follow(g.player)

	// Initialise the renderers
	g.sky = render.NewSky()
	return &g
}

// Destroy frees all resources allocated by the game state.
func (g *Game) Destroy() {
	g.sky.Destroy()
}

// HandleEvent processes a user input event.
func (g *Game) HandleEvent(evt sdl.Event) {
	g.playerCtrl.HandleEvent(evt)
}

// Update advances the game state. It's called at a fixed time step, in order to
// simplify some of the mechanics of the code.
func (g *Game) Update() {
	g.playerCtrl.Update(g.player)
	g.camera.Follow(g.player)

	if g.playerCtrl.IsKeyDown[sdl.SCANCODE_UP] {
		g.worldTime += 0.005
	} else if g.playerCtrl.IsKeyDown[sdl.SCANCODE_DOWN] {
		g.worldTime -= 0.005
	}
}

// Render draws the game to the screen. It's called at a variable time step
// (basically as fast as possible).
func (g *Game) Render() {
	g.sky.Render(render.SkyRenderInfo{
		WorldTime:    g.worldTime,
		Camera:       g.camera,
		RenderRadius: 16,
		LookDir:      g.player.Sight()})
}

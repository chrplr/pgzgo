// Package pgzgo is a small, Pygame-Zero-style harness on top of go-sdl3. It
// removes the boilerplate that every game port repeats: SDL/library
// init-and-teardown, a fixed-step game loop with an FPS cap, window/renderer
// creation, an image cache with drawing helpers, a mixer wrapper, and a
// keyboard/gamepad snapshot layer.
//
// A game supplies its own assets as embedded filesystems (the //go:embed
// directives must live in the game package, since embed can only reach files
// under the importing package's directory) plus two callbacks:
//
//	//go:embed images
//	var imagesFS embed.FS
//	//go:embed sounds music
//	var audioFS embed.FS
//
//	func main() {
//	    app, err := pgzgo.New(pgzgo.Config{
//	        Title: "Kinetix", Width: 640, Height: 640,
//	        Images: imagesFS, Audio: audioFS,
//	    })
//	    if err != nil { panic(err) }
//	    defer app.Close()
//	    setup(app)                 // build your game
//	    app.Loop(update, draw)     // run until quit
//	}
//
// New() performs all initialisation and returns a ready App (Screen, Audio,
// Keyboard, Gamepad). This two-step New/Loop split lets a game run its logic
// headlessly (e.g. a -selftest mode) without entering the loop.
package pgzgo

import (
	"io/fs"

	"github.com/Zyko0/go-sdl3/bin/binimg"
	"github.com/Zyko0/go-sdl3/bin/binmix"
	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
)

// Config describes a game window and where to find its assets.
type Config struct {
	Title  string
	Width  int
	Height int
	FPS    int // frames per second cap; 0 means 60

	// Images is a filesystem containing an "images/" directory of PNGs,
	// typically an embed.FS declared in the game package. May be nil.
	Images fs.FS
	// Audio is a filesystem containing a "sounds/" directory of .ogg effects
	// and an optional "music/" directory of looping .ogg tracks. May be nil.
	Audio fs.FS

	// QuitOnEscape ends the loop when Escape is pressed (default true). Set to
	// false for games that use Escape in-game.
	QuitOnEscape *bool

	// NearestScaling loads textures with nearest-neighbour scaling instead of
	// the default linear filtering — the crisp look of pixel-art games that
	// scale their sprites up (the equivalent of pygame.transform.scale).
	NearestScaling bool
}

// App owns the SDL window, renderer and the sub-systems a game draws through.
type App struct {
	Screen   *Screen
	Audio    *Audio
	Keyboard *Keyboard
	Gamepad  *Gamepad

	// Frame counts elapsed frames; Dt is the last frame's duration in seconds.
	Frame int
	Dt    float64

	window   *sdl.Window
	renderer *sdl.Renderer
	fps      int
	quitOnEsc bool
	quit     bool
	loaders  []interface{ Unload() }
}

// New initialises SDL and its helper libraries, opens the window, and builds the
// Screen, Audio, Keyboard and Gamepad sub-systems. Call Close (usually deferred)
// to tear everything down.
func New(cfg Config) (*App, error) {
	a := &App{
		fps:       cfg.FPS,
		quitOnEsc: cfg.QuitOnEscape == nil || *cfg.QuitOnEscape,
	}
	if a.fps <= 0 {
		a.fps = 60
	}

	// Load the bundled SDL, SDL_image and SDL_mixer shared libraries.
	a.loaders = []interface{ Unload() }{
		binsdl.Load(), binimg.Load(), binmix.Load(),
	}

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		a.Close()
		return nil, err
	}

	window, renderer, err := sdl.CreateWindowAndRenderer(
		cfg.Title, cfg.Width, cfg.Height, 0)
	if err != nil {
		a.Close()
		return nil, err
	}
	a.window = window
	a.renderer = renderer

	a.Screen = newScreen(renderer, cfg.Images, cfg.NearestScaling)
	a.Audio = newAudio(cfg.Audio)
	a.Keyboard = &Keyboard{}
	a.Gamepad = newGamepad()

	return a, nil
}

// Renderer exposes the underlying SDL renderer for advanced drawing.
func (a *App) Renderer() *sdl.Renderer { return a.renderer }

// Quit requests that the game loop end after the current frame.
func (a *App) Quit() { a.quit = true }

// Loop runs the game loop until a quit is requested (window close, Escape when
// enabled, the gamepad Start button, or App.Quit).
//
// The game logic advances at a fixed step, decoupled from how often this callback
// fires. On native, frameSleep paces the callback to the step, so it runs exactly
// one update per iteration (as it always has). In the browser, frameSleep is a
// no-op and the callback is driven by requestAnimationFrame at the display's
// refresh rate — which may be 120 Hz, 144 Hz, etc. Without decoupling, a game that
// moves a fixed amount per update (as these ports do; none scale motion by Dt)
// would run proportionally too fast on a high-refresh display. A time accumulator
// fixes this: it runs however many fixed-step updates the elapsed wall-clock time
// calls for (usually one, sometimes zero on a fast display, rarely more if a frame
// stalled), so the logic rate stays constant on every monitor.
//
// The fixed step is the integer frameMillis, so the average logic rate matches the
// long-standing native rate exactly (e.g. 60 FPS truncates to 16 ms => ~62.5 Hz).
func (a *App) Loop(update, draw func(*App)) {
	frameMillis := uint64(1000 / a.fps)
	step := float64(frameMillis) // fixed logic step, in milliseconds
	const maxCatchUp = 5         // cap updates per callback to avoid a spiral of death

	last := sdl.Ticks()
	var accumulator float64 // unspent wall-clock time, in milliseconds

	sdl.RunLoop(func() error {
		frameStart := sdl.Ticks()

		elapsed := float64(frameStart - last)
		last = frameStart
		if elapsed > maxCatchUp*step { // clamp so a stalled frame can't spiral
			elapsed = maxCatchUp * step
		}
		accumulator += elapsed

		// Nothing is due yet (a callback arrived sooner than the fixed step, e.g.
		// a high-refresh display). Skip input, update and draw entirely — polling
		// input now would roll a key press into "prev" before any update saw it,
		// dropping the rising edge that Keyboard.Pressed reports.
		if accumulator < step {
			return nil
		}

		a.Keyboard.beginFrame()

		var event sdl.Event
		for sdl.PollEvent(&event) {
			if event.Type == sdl.EVENT_QUIT {
				return sdl.EndLoop
			}
			if a.quitOnEsc && event.Type == sdl.EVENT_KEY_DOWN &&
				event.KeyboardEvent().Scancode == sdl.SCANCODE_ESCAPE {
				return sdl.EndLoop
			}
			a.Keyboard.handleEvent(&event)
			a.Gamepad.handleEvent(&event)
		}

		a.Keyboard.refresh()
		a.Gamepad.refresh()
		if a.Gamepad.StartPressed() || a.quit {
			return sdl.EndLoop
		}

		a.Dt = step / 1000.0
		for i := 0; accumulator >= step && i < maxCatchUp; i++ {
			accumulator -= step
			a.Frame++
			if update != nil {
				update(a)
			}
		}

		a.renderer.SetDrawColor(0, 0, 0, 255)
		a.renderer.Clear()
		if draw != nil {
			draw(a)
		}
		a.renderer.Present()

		if elapsed := sdl.Ticks() - frameStart; elapsed < frameMillis {
			frameSleep(frameMillis - elapsed)
		}
		return nil
	})
}

// Close tears down every sub-system and unloads the SDL libraries. It is safe to
// call even after a partially failed New.
func (a *App) Close() {
	if a.Screen != nil {
		a.Screen.Destroy()
	}
	if a.Audio != nil {
		a.Audio.Destroy()
	}
	if a.Gamepad != nil {
		a.Gamepad.close()
	}
	if a.renderer != nil {
		a.renderer.Destroy()
	}
	if a.window != nil {
		a.window.Destroy()
	}
	sdl.Quit()
	// Unload libraries in reverse order of loading.
	for i := len(a.loaders) - 1; i >= 0; i-- {
		a.loaders[i].Unload()
	}
	a.loaders = nil
}

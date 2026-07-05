# pgzgo

A tiny [Pygame Zero](https://pygame-zero.readthedocs.io/)-style harness for making
games in Go on top of [go-sdl3](https://github.com/Zyko0/go-sdl3).

It removes the boilerplate every small SDL game repeats — library load/teardown, a
fixed-step game loop with an FPS cap, window/renderer creation, an image cache with
drawing helpers, a mixer wrapper, and a keyboard/gamepad snapshot layer — so you can
start at the interesting part: your game.

```go
app, _ := pgzgo.New(pgzgo.Config{Title: "Hello", Width: 640, Height: 480, Images: imagesFS})
defer app.Close()
app.Loop(update, draw) // update(app) and draw(app), called every frame
```

pgzgo was extracted from a set of Go ports of the *Code the Classics*
games (see [1](https://github.com/chrplr/Code-the-Classics-Vol1-go/)
and [2](https://github.com/chrplr/Code-the-Classics-Vol2-go/)), where
the same ~350 lines of harness code had been written eight
times. Every game now runs on it.

## Requirements

- Go 1.26+ (as declared in `go.mod`; lower it there if you need to support older toolchains)
- [github.com/Zyko0/go-sdl3](https://github.com/Zyko0/go-sdl3) `v0.1.1` (pulled in
  automatically). Its `bin/*` packages bundle the SDL3, SDL3_image and SDL3_mixer
  shared libraries, so there is no separate system dependency to install.

## Install

```sh
go get github.com/chrplr/pgzgo
```

## Quick start

A complete, self-contained game. Assets are embedded, so the build is a single
binary with no files to ship alongside it.

```go
package main

import (
	"embed"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/pgzgo"
)

//go:embed images
var imagesFS embed.FS // images/player.png, images/title.png, ...

//go:embed sounds music
var audioFS embed.FS // sounds/jump0.ogg, music/theme.ogg, ...

var app *pgzgo.App
var x, y float64

func update(a *pgzgo.App) {
	if a.Keyboard.Held(sdl.SCANCODE_RIGHT) {
		x += 200 * a.Dt // a.Dt is the frame duration in seconds
	}
	if a.Keyboard.Pressed(sdl.SCANCODE_SPACE) {
		a.Audio.Play("jump0")
	}
}

func draw(a *pgzgo.App) {
	a.Screen.Fill(30, 30, 40)
	a.Screen.Blit("player", x, y)
}

func main() {
	a, err := pgzgo.New(pgzgo.Config{
		Title:  "Quick Start",
		Width:  640,
		Height: 480,
		Images: imagesFS,
		Audio:  audioFS,
	})
	if err != nil {
		panic(err)
	}
	defer a.Close()

	app = a
	a.Audio.PlayMusic("theme", 0.5)
	a.Loop(update, draw)
}
```

## The one rule about assets

Go's `//go:embed` can only see files **under the package that declares it**. The
harness therefore cannot embed your assets for you — you declare the `embed.FS` in
your own package and hand it to `Config`:

```
your-game/
├── main.go
├── images/            # embed.FS  → Config.Images
│   ├── player.png
│   └── title.png
└── sounds/            # embed.FS  → Config.Audio (with music/ alongside)
    └── jump0.ogg
```

- `Config.Images` is a filesystem containing an `images/` directory of PNGs.
  `Screen.Blit("player", …)` loads `images/player.png`.
- `Config.Audio` is a filesystem containing a `sounds/` directory of `.ogg` effects
  and an optional `music/` directory of looping `.ogg` tracks. Both are preloaded by
  base name: `sounds/jump0.ogg` → `Audio.Play("jump0")`, `music/theme.ogg` →
  `Audio.PlayMusic("theme", vol)`.

Everything runs from an in-memory copy, so a built binary needs no asset files at run
time.

## Lifecycle: New / Loop / Close

`New` does all initialisation and returns a ready `App`. `Loop` runs the game loop
until quit. `Close` (usually deferred) tears everything down.

The split matters: because `New` gives you a fully working `App` *without* entering
the loop, you can drive your game headlessly — a `-selftest` that steps the logic and
exits, useful for CI:

```go
a, _ := pgzgo.New(cfg)
defer a.Close()
if *selftest {
	runHeadlessChecks(a) // build your game, step Update() N times, assert, return
	return
}
a.Loop(update, draw)
```

Each frame `Loop` polls events, refreshes the keyboard and gamepad snapshots, clears
the screen, then calls `update(app)` and `draw(app)`. It quits on the window close
button, the gamepad **Start** button, `App.Quit()`, and (by default) the **Escape**
key.

## Config

| Field            | Meaning                                                             |
|------------------|---------------------------------------------------------------------|
| `Title`          | Window title.                                                       |
| `Width`, `Height`| Window size in pixels.                                              |
| `FPS`            | Frame-rate cap. `0` means 60.                                        |
| `Images`         | `fs.FS` with an `images/` dir of PNGs. May be nil.                  |
| `Audio`          | `fs.FS` with `sounds/` and optional `music/`. May be nil (see below).|
| `QuitOnEscape`   | `*bool`; Escape quits the loop unless set to `false` (games that use Escape in-game). |
| `NearestScaling` | Load textures with nearest-neighbour scaling (crisp pixel art).     |

## The App and its sub-systems

`App` exposes four sub-systems plus per-frame timing:

- **`app.Screen`** — the drawing surface (Pygame Zero's `screen`):
  `Blit`, `BlitCentred`, `BlitScaled`, `BlitTile`, `Size`, `Texture`, `LoadTexture`
  (from any `fs.FS`, for tilesets/atlases), `Fill`, `FillRect`, `FillRectAlpha`,
  `FillPolygon`, `SetClip`/`ClearClip`, and sprite-font `DrawText`/`TextWidth`.
- **`app.Audio`** — the mixer: `Play(name)` (exact), `PlaySound(name, count)` (random
  variant `name0..name(count-1)`), `PlayMusic(name, volume)`, `StopMusic`.
- **`app.Keyboard`** — `Held(scancode)` and `Pressed(scancode)` (rising edge).
- **`app.Gamepad`** — `Left/Right/Up/Down`, `Button0/Button1` (+ `…Pressed`),
  `AxisX/AxisY`, `StartPressed`, `Connected`. Best-effort: with no controller every
  query returns a neutral value, so games fall back to the keyboard. Hot-plug is
  handled automatically.
- **`app.Dt`** (seconds since last frame, clamped) and **`app.Frame`** (counter).

### Text

`DrawText` takes a `Font`, which maps a rune to a glyph image name. The zero `Font`
uses `"<Prefix>%03d"`; supply a `Name` closure for other conventions (a different
digit format, or substituting a controller-button image for `%`):

```go
font := pgzgo.Font{Space: 22, GapX: -6, Name: func(r rune) string {
	if r == '%' { return "button_a" }
	return fmt.Sprintf("hud%03d", int(r))
}}
app.Screen.DrawText("SCORE %", 20, 10, pgzgo.AlignLeft, font)
```

## Two ways to use the Screen

Games extend the harness at whatever level they need:

**Alias** — when the library covers your needs, the library type *is* your game type:

```go
type Assets = pgzgo.Screen
```

**Embed** — when you need extra draw helpers, embed the Screen so its methods promote
and add your own (which shadow by name and delegate):

```go
type Assets struct {
	*pgzgo.Screen
	terrain image.Image
}

func (a *Assets) CheckTerrain(x, y int) bool { /* game-specific */ }
```

## Opting out of a sub-system

Audio is the common case. If your game needs bespoke audio orchestration (looping
crowd/engine layers, cross-fades), leave `Config.Audio` **nil** — the harness will
not initialise a mixer — and manage your own alongside the harness Screen/Keyboard/
Gamepad. `New` skips mixer setup entirely when `Config.Audio` is nil, so there is no
conflict with a mixer you create yourself.

## WebAssembly (browser)

pgzgo also targets the browser (`GOOS=js GOARCH=wasm`). The wasm build needs the
[go-sdl3-wasm](https://github.com/chrplr/go-sdl3-wasm) fork of go-sdl3, which supplies
the js/wasm bindings, plus its `wasmsdl` bundler. A game redirects go-sdl3 to the fork
for the browser build only (a `go mod edit -replace` that is never committed), so
native `go build` / `go get` are unaffected.

Graphics, keyboard and **audio** all work in the browser — SDL3_mixer plays through
the Emscripten Web Audio backend (the same `Audio.Play` / `PlayMusic` API). Because of
the browser autoplay policy the audio context starts suspended and resumes on the first
user gesture (typically the keypress that starts the game), so title music is silent
until then. Gamepad input is not wired up for wasm yet.

The eight *Code the Classics* ports run this way — playable in-browser with sound; see
their repositories for the GitHub Pages setup.

## License

GNU GPL v3

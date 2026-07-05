# pgzgo — notes for Claude

A small Pygame-Zero-style harness over [go-sdl3](https://github.com/Zyko0/go-sdl3).
Games call `pgzgo.New(Config)` → get an `App` (Screen/Audio/Keyboard/Gamepad) →
`app.Loop(update, draw)` → `app.Close()`.

## Dual target: native and js/wasm

The package compiles for native **and** the browser (`GOOS=js GOARCH=wasm`). The
native path must stay behaviourally unchanged; browser differences live behind
build tags in paired files:

- `platform_notjs.go` / `platform_js.go` — `audioSupported()`, `gamepadSupported()`,
  `frameSleep()`. On js: gamepad off, `frameSleep` is a no-op (rAF paces the loop;
  `SDL_Delay` would freeze the single browser thread), audio **on** (since v0.3.0).
- `screen_notjs.go` / `screen_js.go` — `decodeTexture()`. Native decodes via
  SDL_image; js decodes with Go's own `image/png` and uploads pixels (SDL_image's
  js bindings are stubbed).
- `keyboard_notjs.go` / `keyboard_js.go` — native polls `GetKeyboardState` each
  frame; js is event-driven (`beginFrame`/`handleEvent`), since there is no
  `GetKeyboardState` binding on js. `app.go`'s loop calls all three hooks.

Always verify both: `go build ./... && go vet ./... && GOOS=js GOARCH=wasm go build ./...`

## Assets & audio

- `//go:embed` can only reach files under the *importing* package, so **games own
  the `embed.FS`** and pass them via `Config.Images` / `Config.Audio`.
- `Config.Audio` non-nil → pgzgo runs the mixer (preloads `sounds/*.ogg` and
  looping `music/*.ogg`). `Config.Audio: nil` → opt out; a few games keep a bespoke
  `audio.go` instead.

## Running in the browser

pgzgo alone does **not** produce a browser build — go-sdl3's upstream js bindings
are incomplete. The games' Pages workflow points go-sdl3 at the
[chrplr/go-sdl3-wasm](https://github.com/chrplr/go-sdl3-wasm) fork (via a CI-only
`go mod edit -replace`, never committed) and bundles with the fork's `wasmsdl`
tool. Audio's `AudioContext` starts suspended and SDL's Emscripten backend resumes
it on the first user gesture (the keypress that starts the game), so title music is
silent until then.

## Versions

- `v0.2.0` — added the js/wasm target.
- `v0.3.0` — enabled audio on js/wasm (needs the fork's mixer bindings at build
  time). Tag a new version after changes; games pin an exact `pgzgo` version.

See [[TODO.md]] for the outstanding fork-upstreaming item.

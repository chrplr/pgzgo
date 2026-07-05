//go:build js && wasm

package pgzgo

// audioSupported is true in the browser: SDL3_mixer, its OGG decoder and the
// Web Audio backend are all in the Emscripten build, and go-sdl3-wasm now has
// the js bindings for the mixer playback path. The AudioContext starts
// suspended and SDL's Emscripten backend resumes it on the first user gesture
// (the keypress that starts the game), so title music is silent until then.
func audioSupported() bool { return true }

// gamepadSupported is false in the browser: the SDL gamepad subsystem is not yet
// wired up for js/wasm. Games fall back to keyboard input.
func gamepadSupported() bool { return false }

// frameSleep is a no-op in the browser. Emscripten drives the loop with
// requestAnimationFrame (see sdl.RunLoop), which already paces frames to the
// display; calling SDL_Delay here would block the single browser thread.
func frameSleep(ms uint64) {}

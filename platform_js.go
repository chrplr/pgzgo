//go:build js && wasm

package pgzgo

// audioSupported is false in the browser: SDL3_mixer's js/wasm bindings are not
// yet implemented, so games run silently rather than crash.
func audioSupported() bool { return false }

// gamepadSupported is false in the browser: the SDL gamepad subsystem is not yet
// wired up for js/wasm. Games fall back to keyboard input.
func gamepadSupported() bool { return false }

// frameSleep is a no-op in the browser. Emscripten drives the loop with
// requestAnimationFrame (see sdl.RunLoop), which already paces frames to the
// display; calling SDL_Delay here would block the single browser thread.
func frameSleep(ms uint64) {}

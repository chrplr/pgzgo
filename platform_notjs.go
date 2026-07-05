//go:build !(js && wasm)

package pgzgo

import "github.com/Zyko0/go-sdl3/sdl"

// audioSupported reports whether SDL3_mixer can be initialised on this platform.
func audioSupported() bool { return true }

// gamepadSupported reports whether the SDL gamepad subsystem is available.
func gamepadSupported() bool { return true }

// frameSleep caps the frame rate by sleeping the given number of milliseconds.
// On native builds this is SDL_Delay; the browser build paces frames via
// requestAnimationFrame instead, so there it is a no-op (see platform_js.go).
func frameSleep(ms uint64) { sdl.Delay(uint32(ms)) }

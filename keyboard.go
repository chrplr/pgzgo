package pgzgo

import "github.com/Zyko0/go-sdl3/sdl"

// Keyboard holds this frame's and the previous frame's keyboard snapshots, so it
// can answer both "held" and "just pressed" (rising-edge) queries.
//
// Two mechanisms feed it, chosen at build time. Native builds poll SDL's whole
// keyboard state each frame (refresh, in keyboard_notjs.go). The browser build has
// no SDL_GetKeyboardState binding, so it tracks state from key up/down events
// instead (beginFrame + handleEvent, in keyboard_js.go). App.Loop calls all three
// hooks every frame; the platform that doesn't use one implements it as a no-op.
type Keyboard struct {
	keys []bool
	prev []bool
}

// Held reports whether the key with the given scancode is currently down.
func (k *Keyboard) Held(sc sdl.Scancode) bool {
	return k.keys != nil && int(sc) < len(k.keys) && k.keys[sc]
}

// Pressed reports the rising edge of a key: down this frame, up last frame.
func (k *Keyboard) Pressed(sc sdl.Scancode) bool {
	return k.Held(sc) && (k.prev == nil || int(sc) >= len(k.prev) || !k.prev[sc])
}

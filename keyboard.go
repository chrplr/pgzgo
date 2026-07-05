package pgzgo

import "github.com/Zyko0/go-sdl3/sdl"

// Keyboard holds this frame's and the previous frame's keyboard snapshots, so it
// can answer both "held" and "just pressed" (rising-edge) queries. App.Loop
// refreshes it once per frame.
type Keyboard struct {
	keys []bool
	prev []bool
}

func (k *Keyboard) refresh() {
	current := sdl.GetKeyboardState()
	if current == nil {
		return
	}
	if len(k.prev) != len(current) {
		k.prev = make([]bool, len(current))
	}
	if len(k.keys) != len(current) {
		k.keys = make([]bool, len(current))
	}
	copy(k.prev, k.keys)
	copy(k.keys, current)
}

// Held reports whether the key with the given scancode is currently down.
func (k *Keyboard) Held(sc sdl.Scancode) bool {
	return k.keys != nil && int(sc) < len(k.keys) && k.keys[sc]
}

// Pressed reports the rising edge of a key: down this frame, up last frame.
func (k *Keyboard) Pressed(sc sdl.Scancode) bool {
	return k.Held(sc) && (k.prev == nil || int(sc) >= len(k.prev) || !k.prev[sc])
}

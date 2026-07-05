//go:build !(js && wasm)

package pgzgo

import "github.com/Zyko0/go-sdl3/sdl"

// beginFrame is a no-op on native builds; state is captured wholesale in refresh.
func (k *Keyboard) beginFrame() {}

// handleEvent is a no-op on native builds; refresh reads the full keyboard state.
func (k *Keyboard) handleEvent(evt *sdl.Event) {}

// refresh snapshots SDL's whole keyboard state, keeping the previous frame for
// rising-edge detection. App.Loop calls it once per frame.
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

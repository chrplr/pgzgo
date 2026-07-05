//go:build js && wasm

package pgzgo

import "github.com/Zyko0/go-sdl3/sdl"

// In the browser there is no SDL_GetKeyboardState binding, so the keyboard is
// tracked from discrete key up/down events delivered by SDL_PollEvent. beginFrame
// rolls this frame's state into prev (for rising-edge detection) and handleEvent
// applies each event; refresh has nothing to do.

func (k *Keyboard) ensure() {
	if k.keys == nil {
		k.keys = make([]bool, sdl.SCANCODE_COUNT)
		k.prev = make([]bool, sdl.SCANCODE_COUNT)
	}
}

// beginFrame copies the current key state into prev, so keys pressed during this
// frame's events register as rising edges. App.Loop calls it before polling.
func (k *Keyboard) beginFrame() {
	k.ensure()
	copy(k.prev, k.keys)
}

// handleEvent updates the held state from a key up/down event.
func (k *Keyboard) handleEvent(evt *sdl.Event) {
	switch evt.Type {
	case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
		k.ensure()
		sc := evt.KeyboardEvent().Scancode
		if int(sc) < len(k.keys) {
			k.keys[sc] = evt.Type == sdl.EVENT_KEY_DOWN
		}
	}
}

// refresh is a no-op in the browser; state comes from events, not polling.
func (k *Keyboard) refresh() {}

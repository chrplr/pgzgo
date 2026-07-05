package pgzgo

import "github.com/Zyko0/go-sdl3/sdl"

// Gamepad wraps a single active controller. SDL3's "gamepad" API applies a
// controller-mapping database, so buttons and axes are named rather than raw
// indices — the equivalent of a mapped Pygame joystick. Button state is
// snapshotted each frame so both "held" and "just pressed" (edge) queries work.
//
// The harness reads the South/East face buttons (Xbox A/B), the d-pad, the left
// analogue stick, and Start (which App.Loop treats as quit).
type Gamepad struct {
	pad  *sdl.Gamepad
	id   sdl.JoystickID
	cur  map[sdl.GamepadButton]bool
	prev map[sdl.GamepadButton]bool
}

var trackedButtons = []sdl.GamepadButton{
	sdl.GAMEPAD_BUTTON_SOUTH, sdl.GAMEPAD_BUTTON_EAST,
	sdl.GAMEPAD_BUTTON_DPAD_LEFT, sdl.GAMEPAD_BUTTON_DPAD_RIGHT,
	sdl.GAMEPAD_BUTTON_DPAD_UP, sdl.GAMEPAD_BUTTON_DPAD_DOWN,
	sdl.GAMEPAD_BUTTON_START,
}

// newGamepad opens the gamepad subsystem and the first available controller.
// It is best-effort: with no controller (or on a headless machine) it leaves
// the pad nil and every query returns a neutral value, so games fall back to
// the keyboard.
func newGamepad() *Gamepad {
	g := &Gamepad{
		cur:  map[sdl.GamepadButton]bool{},
		prev: map[sdl.GamepadButton]bool{},
	}
	if err := sdl.InitSubSystem(sdl.INIT_GAMEPAD); err != nil {
		return g
	}
	ids, err := sdl.GetGamepads()
	if err != nil {
		return g
	}
	for _, id := range ids {
		if g.open(id) {
			break
		}
	}
	return g
}

func (g *Gamepad) open(id sdl.JoystickID) bool {
	pad, err := id.OpenGamepad()
	if err != nil || pad == nil {
		return false
	}
	g.pad = pad
	g.id = id
	return true
}

// handleEvent processes hot-plug add/remove events so a controller can be
// connected or disconnected while the game runs. App.Loop calls it per event.
func (g *Gamepad) handleEvent(evt *sdl.Event) {
	switch evt.Type {
	case sdl.EVENT_GAMEPAD_ADDED:
		if g.pad == nil {
			g.open(evt.GamepadDeviceEvent().Which)
		}
	case sdl.EVENT_GAMEPAD_REMOVED:
		if g.pad != nil && evt.GamepadDeviceEvent().Which == g.id {
			g.pad.Close()
			g.pad = nil
		}
	}
}

func (g *Gamepad) connected() bool { return g.pad != nil && g.pad.Connected() }

func (g *Gamepad) live(b sdl.GamepadButton) bool {
	if !g.connected() {
		return false
	}
	return g.pad.Button(b)
}

// refresh snapshots the tracked buttons; App.Loop calls it once per frame.
func (g *Gamepad) refresh() {
	for _, b := range trackedButtons {
		g.prev[b] = g.cur[b]
		g.cur[b] = g.live(b)
	}
}

func (g *Gamepad) held(b sdl.GamepadButton) bool    { return g.cur[b] }
func (g *Gamepad) pressed(b sdl.GamepadButton) bool { return g.cur[b] && !g.prev[b] }

// axis returns an analogue axis normalised to roughly [-1, 1] (read live — axes
// don't need edge detection).
func (g *Gamepad) axis(a sdl.GamepadAxis) float64 {
	if !g.connected() {
		return 0
	}
	return float64(g.pad.Axis(a)) / 32767.0
}

// Connected reports whether a controller is currently attached.
func (g *Gamepad) Connected() bool { return g.connected() }

// Directional d-pad queries (held).
func (g *Gamepad) Left() bool  { return g.held(sdl.GAMEPAD_BUTTON_DPAD_LEFT) }
func (g *Gamepad) Right() bool { return g.held(sdl.GAMEPAD_BUTTON_DPAD_RIGHT) }
func (g *Gamepad) Up() bool    { return g.held(sdl.GAMEPAD_BUTTON_DPAD_UP) }
func (g *Gamepad) Down() bool  { return g.held(sdl.GAMEPAD_BUTTON_DPAD_DOWN) }

// Face-button queries. Button0 is South (Xbox A), Button1 is East (Xbox B).
func (g *Gamepad) Button0() bool        { return g.held(sdl.GAMEPAD_BUTTON_SOUTH) }
func (g *Gamepad) Button1() bool        { return g.held(sdl.GAMEPAD_BUTTON_EAST) }
func (g *Gamepad) Button0Pressed() bool { return g.pressed(sdl.GAMEPAD_BUTTON_SOUTH) }
func (g *Gamepad) Button1Pressed() bool { return g.pressed(sdl.GAMEPAD_BUTTON_EAST) }

// StartPressed reports the rising edge of the Start button (App.Loop quits on it).
func (g *Gamepad) StartPressed() bool { return g.pressed(sdl.GAMEPAD_BUTTON_START) }

// Left analogue stick axes, each normalised to roughly [-1, 1].
func (g *Gamepad) AxisX() float64 { return g.axis(sdl.GAMEPAD_AXIS_LEFTX) }
func (g *Gamepad) AxisY() float64 { return g.axis(sdl.GAMEPAD_AXIS_LEFTY) }

func (g *Gamepad) close() {
	if g.pad != nil {
		g.pad.Close()
		g.pad = nil
	}
}

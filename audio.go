package pgzgo

import (
	"io/fs"
	"math/rand"
	"path"
	"strconv"

	"github.com/Zyko0/go-sdl3/mixer"
	"github.com/Zyko0/go-sdl3/sdl"
)

// Audio wraps SDL3_mixer. Every operation is best-effort: if the mixer fails to
// initialise (e.g. a headless machine) the methods simply do nothing, so games
// run unchanged without sound.
type Audio struct {
	fsys   fs.FS
	mixer  *mixer.Mixer
	sounds map[string]*mixer.Audio

	music        map[string]*mixer.Track
	currentMusic *mixer.Track
}

func newAudio(fsys fs.FS) *Audio {
	a := &Audio{
		fsys:   fsys,
		sounds: make(map[string]*mixer.Audio),
		music:  make(map[string]*mixer.Track),
	}
	if fsys == nil || !audioSupported() {
		// audioSupported() is false on js/wasm, where SDL3_mixer's bindings are
		// not yet implemented; the game then runs silently with a nil mixer.
		return a
	}
	if err := mixer.Init(); err != nil {
		return a
	}
	m, err := mixer.CreateMixerDevice(sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK, nil)
	if err != nil {
		return a
	}
	a.mixer = m

	// Preload every embedded sounds/*.ogg effect, keyed by base name.
	if entries, err := fs.ReadDir(fsys, "sounds"); err == nil {
		for _, e := range entries {
			fname := e.Name()
			if path.Ext(fname) != ".ogg" {
				continue
			}
			if snd := a.load("sounds/"+fname, true); snd != nil {
				a.sounds[fname[:len(fname)-len(".ogg")]] = snd
			}
		}
	}

	// Preload every embedded music/*.ogg as a looping track, keyed by base name.
	if entries, err := fs.ReadDir(fsys, "music"); err == nil {
		for _, e := range entries {
			fname := e.Name()
			if path.Ext(fname) != ".ogg" {
				continue
			}
			if t := a.loopingTrack(m, "music/"+fname); t != nil {
				a.music[fname[:len(fname)-len(".ogg")]] = t
			}
		}
	}
	return a
}

// load decodes an embedded audio file into an in-memory Audio via an SDL
// IOStream. predecode fully decodes short effects up front.
func (a *Audio) load(p string, predecode bool) *mixer.Audio {
	data, err := fs.ReadFile(a.fsys, p)
	if err != nil {
		return nil
	}
	stream, err := sdl.IOFromConstMem(data)
	if err != nil {
		return nil
	}
	snd, err := a.mixer.LoadAudio_IO(stream, predecode, true) // closeio
	if err != nil {
		return nil
	}
	return snd
}

func (a *Audio) loopingTrack(m *mixer.Mixer, p string) *mixer.Track {
	audio := a.load(p, false)
	if audio == nil {
		return nil
	}
	t, err := m.CreateTrack()
	if err != nil {
		return nil
	}
	t.SetAudio(audio)
	t.SetLoops(-1)
	return t
}

// PlaySound plays one of a family of sound variants. With count > 1 it picks a
// random variant name0 .. name(count-1); with count <= 1 it plays name0. This
// mirrors Pygame Zero's convention where e.g. sounds.hit0 .. hit3 are variants.
func (a *Audio) PlaySound(name string, count int) {
	if a.mixer == nil {
		return
	}
	variant := name + "0"
	if count > 1 {
		variant = name + strconv.Itoa(rand.Intn(count))
	}
	if snd, ok := a.sounds[variant]; ok && snd != nil {
		a.mixer.PlayAudio(snd)
	}
}

// Play plays a single sound effect by its exact name (sounds/<name>.ogg), with
// no variant suffix — the equivalent of Pygame Zero's sounds.<name>.play().
func (a *Audio) Play(name string) {
	if a.mixer == nil {
		return
	}
	if snd, ok := a.sounds[name]; ok && snd != nil {
		a.mixer.PlayAudio(snd)
	}
}

// PlayMusic switches to the named looping music track (music/<name>.ogg) at the
// given gain (0..1). Any current track is stopped first.
func (a *Audio) PlayMusic(name string, volume float32) {
	a.StopMusic()
	if t, ok := a.music[name]; ok && t != nil {
		t.SetGain(volume)
		t.Play(0)
		a.currentMusic = t
	}
}

// StopMusic stops the current looping track, if any.
func (a *Audio) StopMusic() {
	if a.currentMusic != nil {
		a.currentMusic.Stop(0)
		a.currentMusic = nil
	}
}

// SoundCount reports how many sound effects were successfully loaded. Useful in
// headless self-tests to confirm the embedded audio decoded.
func (a *Audio) SoundCount() int { return len(a.sounds) }

// Destroy releases the mixer device.
func (a *Audio) Destroy() {
	if a.mixer != nil {
		a.mixer.Destroy()
		a.mixer = nil
	}
}

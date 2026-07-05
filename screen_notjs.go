//go:build !(js && wasm)

package pgzgo

import (
	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
)

// decodeTexture turns encoded image bytes into a texture using SDL_image, via an
// in-memory SDL IOStream so no file is touched at run time.
func decodeTexture(renderer *sdl.Renderer, data []byte) *sdl.Texture {
	stream, err := sdl.IOFromConstMem(data)
	if err != nil {
		return nil
	}
	tex, err := img.LoadTextureIO(renderer, stream, true) // closeio: frees the stream
	if err != nil {
		return nil
	}
	return tex
}

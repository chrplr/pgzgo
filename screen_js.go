//go:build js && wasm

package pgzgo

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"

	"github.com/Zyko0/go-sdl3/sdl"
)

// decodeTexture turns encoded image bytes into a texture without SDL_image, whose
// js/wasm bindings are still stubbed. It decodes with Go's own image codecs, copies
// the pixels into a tightly-packed RGBA buffer, and uploads them to a static
// texture. PNG covers every embedded asset; JPEG is registered for completeness.
func decodeTexture(renderer *sdl.Renderer, data []byte) *sdl.Texture {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	rgba, ok := src.(*image.RGBA)
	if !ok || rgba.Stride != w*4 || !rgba.Rect.Eq(image.Rect(0, 0, w, h)) {
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
		rgba = dst
	}
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA32, sdl.TEXTUREACCESS_STATIC, w, h)
	if err != nil {
		return nil
	}
	tex.Update(nil, rgba.Pix, int32(rgba.Stride))
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	return tex
}

// ensure png stays imported even though image.Decode drives it via the registry.
var _ = png.Decode

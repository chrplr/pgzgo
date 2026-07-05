package pgzgo

import (
	"io/fs"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Screen is the drawing surface: a lazily-populated cache of textures decoded
// from the game's embedded images, plus blit/fill/clip helpers. It is the
// equivalent of Pygame Zero's global `screen`.
type Screen struct {
	renderer *sdl.Renderer
	images   fs.FS
	textures map[string]*sdl.Texture
	nearest  bool
}

func newScreen(renderer *sdl.Renderer, images fs.FS, nearest bool) *Screen {
	return &Screen{
		renderer: renderer,
		images:   images,
		textures: make(map[string]*sdl.Texture),
		nearest:  nearest,
	}
}

// Texture lazily decodes images/<name>.png from the embedded filesystem and
// caches it. A load failure caches (and returns) nil so callers can no-op.
func (s *Screen) Texture(name string) *sdl.Texture {
	if tex, ok := s.textures[name]; ok {
		return tex
	}
	tex := loadTextureFromFS(s.renderer, s.images, "images/"+name+".png")
	if tex != nil && s.nearest {
		tex.SetScaleMode(sdl.SCALEMODE_NEAREST)
	}
	s.textures[name] = tex
	return tex
}

// loadTextureFromFS reads an embedded image and decodes it into a texture. The
// bytes→texture step (decodeTexture) is platform-specific: native builds use
// SDL_image, while the js/wasm build decodes with Go's image/png and uploads the
// pixels directly, so it doesn't depend on the (still-stubbed) SDL_image bindings.
func loadTextureFromFS(renderer *sdl.Renderer, fsys fs.FS, path string) *sdl.Texture {
	if fsys == nil {
		return nil
	}
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil
	}
	return decodeTexture(renderer, data)
}

// LoadTexture decodes and caches a texture from an arbitrary filesystem, keyed by
// its path. Use it for assets that don't live under images/ — tilesets, sprite
// atlases, or a second embedded FS. Returns nil (cached) on failure.
func (s *Screen) LoadTexture(fsys fs.FS, path string) *sdl.Texture {
	if tex, ok := s.textures[path]; ok {
		return tex
	}
	tex := loadTextureFromFS(s.renderer, fsys, path)
	if tex != nil && s.nearest {
		tex.SetScaleMode(sdl.SCALEMODE_NEAREST)
	}
	s.textures[path] = tex
	return tex
}

// Size returns the width and height of a named image (0,0 if it failed to load).
func (s *Screen) Size(name string) (float64, float64) {
	tex := s.Texture(name)
	if tex == nil {
		return 0, 0
	}
	return float64(tex.W), float64(tex.H)
}

// Blit draws an image with its top-left corner at (x, y).
func (s *Screen) Blit(name string, x, y float64) {
	tex := s.Texture(name)
	if tex == nil {
		return
	}
	dst := sdl.FRect{X: float32(x), Y: float32(y), W: float32(tex.W), H: float32(tex.H)}
	s.renderer.RenderTexture(tex, nil, &dst)
}

// BlitCentred draws an image centred on (cx, cy), like a Pygame Zero Actor.
func (s *Screen) BlitCentred(name string, cx, cy float64) {
	w, h := s.Size(name)
	s.Blit(name, cx-w/2, cy-h/2)
}

// BlitScaled draws an image scaled to (w, h) with its top-left corner at (x, y).
func (s *Screen) BlitScaled(name string, x, y, w, h float64) {
	tex := s.Texture(name)
	if tex == nil {
		return
	}
	dst := sdl.FRect{X: float32(x), Y: float32(y), W: float32(w), H: float32(h)}
	s.renderer.RenderTexture(tex, nil, &dst)
}

// BlitTile draws a sw×sh sub-rectangle of a texture at (dx, dy). Useful for
// tilesets and sprite sheets.
func (s *Screen) BlitTile(tex *sdl.Texture, sx, sy, sw, sh, dx, dy float64) {
	if tex == nil {
		return
	}
	src := sdl.FRect{X: float32(sx), Y: float32(sy), W: float32(sw), H: float32(sh)}
	dst := sdl.FRect{X: float32(dx), Y: float32(dy), W: float32(sw), H: float32(sh)}
	s.renderer.RenderTexture(tex, &src, &dst)
}

// Fill clears the whole frame to a solid colour.
func (s *Screen) Fill(r, g, b uint8) {
	s.renderer.SetDrawColor(r, g, b, 255)
	s.renderer.Clear()
}

// FillRect draws a solid rectangle.
func (s *Screen) FillRect(x, y, w, h float64, r, g, b uint8) {
	s.renderer.SetDrawColor(r, g, b, 255)
	rect := sdl.FRect{X: float32(x), Y: float32(y), W: float32(w), H: float32(h)}
	s.renderer.RenderFillRect(&rect)
}

// FillRectAlpha draws a translucent rectangle (e.g. a fade-to-black overlay).
func (s *Screen) FillRectAlpha(x, y, w, h float64, r, g, b, alpha uint8) {
	s.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	s.renderer.SetDrawColor(r, g, b, alpha)
	rect := sdl.FRect{X: float32(x), Y: float32(y), W: float32(w), H: float32(h)}
	s.renderer.RenderFillRect(&rect)
}

// FillPolygon fills a convex polygon (given as screen-space (x, y) pairs) with a
// solid colour, using a triangle fan — the equivalent of pygame.draw.polygon.
func (s *Screen) FillPolygon(points [][2]float64, r, g, b uint8) {
	if len(points) < 3 {
		return
	}
	col := sdl.FColor{R: float32(r) / 255, G: float32(g) / 255, B: float32(b) / 255, A: 1}
	verts := make([]sdl.Vertex, len(points))
	for i, p := range points {
		verts[i] = sdl.Vertex{
			Position: sdl.FPoint{X: float32(p[0]), Y: float32(p[1])},
			Color:    col,
		}
	}
	indices := make([]int32, 0, (len(points)-2)*3)
	for i := 1; i < len(points)-1; i++ {
		indices = append(indices, 0, int32(i), int32(i+1))
	}
	s.renderer.RenderGeometry(nil, verts, indices)
}

// SetClip restricts subsequent drawing to the given rectangle.
func (s *Screen) SetClip(x, y, w, h int32) {
	r := sdl.Rect{X: x, Y: y, W: w, H: h}
	s.renderer.SetClipRect(&r)
}

// ClearClip removes any clipping rectangle.
func (s *Screen) ClearClip() {
	s.renderer.SetClipRect(nil)
}

// Destroy frees every cached texture.
func (s *Screen) Destroy() {
	for _, tex := range s.textures {
		if tex != nil {
			tex.Destroy()
		}
	}
	s.textures = nil
}

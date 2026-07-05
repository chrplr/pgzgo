package pgzgo

import "fmt"

// Align selects horizontal text alignment relative to the given x coordinate.
type Align int

const (
	AlignLeft Align = iota
	AlignCentre
	AlignRight
)

// Font describes a sprite font: each character is its own image, named by a
// convention. These games store a glyph per ASCII code (e.g. "font065" for 'A'),
// so Name maps a rune to an image name; GapX is extra spacing between glyphs and
// Space is the advance width of a space character.
//
// The zero Font is usable: it maps rune r to "<Prefix>%03d" (the most common
// convention here). Provide a custom Name for anything else (a different digit
// format, or button-glyph substitutions like '%' -> "xb_a").
type Font struct {
	Prefix string
	Space  float64
	GapX   float64
	Name   func(r rune) string
}

func (f Font) name(r rune) string {
	if r == ' ' {
		return ""
	}
	if f.Name != nil {
		return f.Name(r)
	}
	return fmt.Sprintf("%s%03d", f.Prefix, int(r))
}

// TextWidth returns the pixel width of text rendered in the given font.
func (s *Screen) TextWidth(text string, f Font) float64 {
	runes := []rune(text)
	total := 0.0
	for _, r := range runes {
		if name := f.name(r); name == "" {
			total += f.Space
		} else {
			w, _ := s.Size(name)
			total += w
		}
	}
	if len(runes) > 1 {
		total += f.GapX * float64(len(runes)-1)
	}
	return total
}

// DrawText draws text using a sprite font with the given horizontal alignment.
func (s *Screen) DrawText(text string, x, y float64, align Align, f Font) {
	switch align {
	case AlignCentre:
		x -= float64(int(s.TextWidth(text, f)) / 2)
	case AlignRight:
		x -= s.TextWidth(text, f)
	}
	for _, r := range text {
		name := f.name(r)
		if name == "" {
			x += f.Space + f.GapX
			continue
		}
		s.Blit(name, x, y)
		w, _ := s.Size(name)
		x += w + f.GapX
	}
}

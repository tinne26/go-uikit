package ui

import (
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font/sfnt"
)

// SanitizeString replaces runes that are not present in the font with a safe fallback.
// This prevents renderer panics on missing glyphs.
func SanitizeString(font *sfnt.Font, s string) string {
	if font == nil || s == "" {
		return s
	}

	var buf sfnt.Buffer
	q, _ := font.GlyphIndex(&buf, '?')
	useQ := q != 0

	out := make([]rune, 0, len(s))
	for len(s) > 0 {
		r, n := utf8.DecodeRuneInString(s)
		s = s[n:]
		if r == utf8.RuneError && n == 1 {
			continue
		}
		gid, err := font.GlyphIndex(&buf, r)
		if err == nil && gid != 0 {
			out = append(out, r)
			continue
		}
		if useQ {
			out = append(out, '?')
		}
	}
	return string(out)
}

// DrawTextSafe draws text while protecting against missing-glyph panics.
// It does not modify renderer configuration (align, size, color), it only sanitizes the string.
func DrawTextSafe(ctx *Context, dst *ebiten.Image, s string, x, y int) {
	if ctx == nil || ctx.Text == nil || dst == nil || s == "" {
		return
	}
	s = SanitizeString(ctx.Theme.Font, s)
	defer func() { _ = recover() }()
	ctx.Text.Draw(dst, s, x, y)
}

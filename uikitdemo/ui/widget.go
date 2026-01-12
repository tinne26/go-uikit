package ui

import "github.com/hajimehoshi/ebiten/v2"

type Widget interface {
	Base() *Base
	Focusable() bool
	// SetFrame sets the widget position and width. Height is derived from the theme and widget state.
	SetFrame(x, y, w int)
	// Measure returns the current widget rectangle (including any extra height such as validation errors).
	Measure() Rect
	Update(ctx *Context)
	Draw(ctx *Context, dst *ebiten.Image)
}

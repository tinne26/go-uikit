package ui

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// LayoutStack is a small helper to lay out a vertical stack of widgets inside a viewport,
// with internal scrolling (wheel + drag/touch).
//
// It is intentionally NOT part of the widget core; it only sets widget frames and toggles
// visibility based on intersection with the viewport.
type LayoutStack struct {
	Viewport Rect

	// Padding inside the viewport before placing widgets.
	PadX int
	PadY int

	// Gap between widgets.
	Gap int

	// ScrollY is the current scroll offset in pixels (positive = content moved up).
	ScrollY int

	// Internal dragging state
	dragging bool
	lastPY   int
}

// NewLayoutStack creates a LayoutStack with sensible defaults based on the theme.
func NewLayoutStack(theme *Theme) *LayoutStack {
	return &LayoutStack{
		PadX: theme.SpaceM,
		PadY: theme.SpaceM,
		Gap:  theme.SpaceS,
	}
}

// UpdateScroll updates the scroll position based on input (wheel + drag) when the pointer is inside the viewport.
func (ls *LayoutStack) UpdateScroll(ctx *Context, contentH int) {
	if ls.Viewport.W <= 0 || ls.Viewport.H <= 0 {
		return
	}

	x, y, down, justDown, justUp, _ := ctx.Pointer()
	inside := ls.Viewport.Contains(x, y)

	// Wheel (desktop)
	_, wy := ebiten.Wheel()
	if wy != 0 && inside {
		step := int(math.Round(float64(ctx.Theme.ControlH) * 0.65))
		if step < 10 {
			step = 10
		}
		ls.ScrollY -= int(math.Round(wy * float64(step)))
	}

	// Drag (mouse/touch)
	if justDown && inside {
		ls.dragging = true
		ls.lastPY = y
	}
	if ls.dragging && down {
		dy := y - ls.lastPY
		ls.ScrollY -= dy
		ls.lastPY = y
	}
	if justUp {
		ls.dragging = false
	}

	ls.ClampScroll(contentH)
}

// ClampScroll clamps ScrollY to [0..max].
func (ls *LayoutStack) ClampScroll(contentH int) {
	max := contentH - ls.Viewport.H
	if max < 0 {
		max = 0
	}
	if ls.ScrollY < 0 {
		ls.ScrollY = 0
	}
	if ls.ScrollY > max {
		ls.ScrollY = max
	}
}

// Apply lays out widgets vertically inside the viewport and returns the full content height.
// Widgets completely outside the viewport are set to Visible=false to avoid drawing and hit-testing.
func (ls *LayoutStack) Apply(ctx *Context, widgets []Widget) int {
	x0 := ls.Viewport.X + ls.PadX
	y0 := ls.Viewport.Y + ls.PadY - ls.ScrollY
	w0 := ls.Viewport.W - ls.PadX*2
	if w0 < 0 {
		w0 = 0
	}

	y := y0
	contentH := ls.PadY * 2

	for _, w := range widgets {
		b := w.Base()
		// keep enabled as-is; only toggle visibility for clipping.
		w.SetFrame(x0, y, w0)
		r := w.Measure()

		// Total content height in scroll space
		contentH += r.H

		// Visibility: hide only when fully outside viewport (cheap clipping).
		// This avoids drawing/hit-testing offscreen widgets. Partial overlap may still draw outside.
		vr := ls.Viewport
		if r.Bottom() < vr.Y || r.Y > vr.Bottom() {
			b.Visible = false
		} else {
			b.Visible = true
		}

		y += r.H + ls.Gap
		contentH += ls.Gap
	}

	// Clamp after layout
	ls.ClampScroll(contentH)
	return contentH
}

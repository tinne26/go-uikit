package ui

import "github.com/hajimehoshi/ebiten/v2"

// Hittable allows a widget to extend its clickable area beyond Base.ControlRect.
type Hittable interface {
	HitTest(ctx *Context, x, y int) bool
}

// OverlayWidget can draw an overlay above all other widgets (e.g. Select dropdown).
type OverlayWidget interface {
	OverlayActive() bool
	DrawOverlay(ctx *Context, dst *ebiten.Image)
}

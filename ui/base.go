package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Base contains shared widget state.
// Height is usually Theme.ControlH, but can be overridden per widget (e.g. TextArea). External layout controls only X/Y/Width.
type Base struct {
	Rect    Rect
	Visible bool
	Enabled bool

	Invalid   bool
	ErrorText string

	// ControlH overrides Theme.ControlH when > 0 (for variable-height controls like TextArea).
	ControlH int

	// internal state (managed by Context)
	hovered bool
	pressed bool
	focused bool

	theme *Theme
}

func NewBase() Base {
	return Base{
		Visible: true,
		Enabled: true,
	}
}

func (b *Base) ControlHeight(theme *Theme) int {
	if b.ControlH > 0 {
		return b.ControlH
	}
	return theme.ControlH
}

func (b *Base) RequiredHeight(theme *Theme) int {
	h := b.ControlHeight(theme)
	if b.Invalid && b.ErrorText != "" {
		// error line height uses theme.ErrorFontPx metrics
		met, _ := MetricsPx(theme.Font, theme.ErrorFontPx)
		h += theme.ErrorGap + met.Height
	}
	return h
}

func (b *Base) ControlRect(theme *Theme) Rect {
	return Rect{X: b.Rect.X, Y: b.Rect.Y, W: b.Rect.W, H: b.ControlHeight(theme)}
}

func (b *Base) ErrorRect(theme *Theme) Rect {
	if !(b.Invalid && b.ErrorText != "") {
		return Rect{}
	}
	met, _ := MetricsPx(theme.Font, theme.ErrorFontPx)
	y := b.Rect.Y + b.ControlHeight(theme) + theme.ErrorGap
	return Rect{X: b.Rect.X, Y: y, W: b.Rect.W, H: met.Height}
}

func (b *Base) SetInvalid(err string) {
	b.Invalid = true
	b.ErrorText = err
}

func (b *Base) ClearInvalid() {
	b.Invalid = false
	b.ErrorText = ""
}

// SetFrame sets the widget position (x,y) and width (w).
// The final height is computed from the theme and the widget state (e.g. error line, variable-height controls).
func (b *Base) SetFrame(theme *Theme, x, y, w int) {
	if w < 0 {
		w = 0
	}
	b.Rect = Rect{X: x, Y: y, W: w, H: b.RequiredHeight(theme)}
}

func (b *Base) Hovered() bool { return b.hovered }
func (b *Base) Pressed() bool { return b.pressed }
func (b *Base) Focused() bool { return b.focused }

func (b *Base) SetEnabled(v bool) { b.Enabled = v }
func (b *Base) SetVisible(v bool) { b.Visible = v }

func (c *Base) Draw(ctx *Context, dst *ebiten.Image) Rect {
	c.theme = ctx.Theme
	if c.Rect.H == 0 {
		c.SetFrame(ctx.Theme, c.Rect.X, c.Rect.Y, c.Rect.W)
	}

	r := c.ControlRect(ctx.Theme)

	// Surface
	bg := ctx.Theme.Surface
	if !c.Enabled {
		bg = ctx.Theme.SurfacePressed
	} else if c.pressed {
		bg = ctx.Theme.SurfacePressed
	} else if c.hovered {
		bg = ctx.Theme.SurfaceHover
	}

	drawRoundedRect(dst, r, ctx.Theme.Radius, bg)

	// Border
	border := ctx.Theme.Border
	if !c.Enabled {
		border = ctx.Theme.Disabled
	}
	if c.Invalid {
		border = ctx.Theme.ErrorBorder
	}

	drawRoundedBorder(dst, r, ctx.Theme.Radius, ctx.Theme.BorderW, border)

	// Focus ring
	if c.focused && c.Enabled {
		drawFocusRing(dst, r, ctx.Theme.Radius, ctx.Theme.FocusRingGap, ctx.Theme.FocusRingW, ctx.Theme.Focus)
	}

	// Error
	err := c.ErrorRect(ctx.Theme)
	if c.Invalid {
		drawErrorText(ctx, dst, err, c.ErrorText)
	}

	return r
}

// SetTheme allows layouts to provide Theme before SetFrame is called.
func (c *Base) SetTheme(theme *Theme) {
	c.theme = theme
}

func (c *Base) Theme() *Theme {
	return c.theme
}

package ui

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

	hovered bool
	pressed bool
	focused bool
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

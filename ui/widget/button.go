package widget

import (
	"image"

	"github.com/erparts/go-uikit/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Button struct {
	base    ui.Base
	label   string
	OnClick func()
}

func NewButton(theme *ui.Theme, label string) *Button {
	cfg := ui.NewWidgetBaseConfig(theme)

	return &Button{
		base:  ui.NewBase(cfg),
		label: label,
	}
}

func (b *Button) Base() *ui.Base {
	return &b.base
}

func (b *Button) Focusable() bool {
	return true
}

func (b *Button) SetFrame(x, y, w int) {
	b.base.SetFrame(x, y, w)
}

func (b *Button) Measure() image.Rectangle {
	return b.base.Rect
}

func (b *Button) SetEnabled(v bool) {
	b.base.SetEnabled(v)
}

func (b *Button) SetVisible(v bool) {
	b.base.SetVisible(v)
}

func (b *Button) SetLabel(s string) { b.label = s }

func (b *Button) HandleEvent(ctx *ui.Context, e ui.Event) {
	if !b.base.IsEnabled() {
		return
	}

	if e.Type == ui.EventClick {
		if b.OnClick != nil {
			b.OnClick()
		}
	}
}

func (b *Button) Update(ctx *ui.Context) {
	if !b.base.IsEnabled() {
		return
	}

	// Keyboard click when focused
	if b.base.IsFocused() && (inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) {
		if b.OnClick != nil {
			b.OnClick()
		}
	}
}

func (b *Button) Draw(ctx *ui.Context, dst *ebiten.Image) {
	r := b.base.Draw(ctx, dst)

	// Centered label
	met, _ := ui.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	textW := ui.MeasureStringPx(ctx.Theme.Font, ctx.Theme.FontPx, b.label)

	tx := r.Min.X + (r.Dx()-textW)/2
	baselineY := r.Min.Y + (r.Dy()-met.Height)/2 + met.Ascent

	col := ctx.Theme.Text
	if !b.base.IsEnabled() {
		col = ctx.Theme.Disabled
	}

	ctx.Text.SetColor(col)
	ctx.Text.SetAlign(0) // Left
	ctx.Text.Draw(dst, b.label, tx, baselineY)
}

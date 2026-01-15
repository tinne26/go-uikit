package widget

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

type Label struct {
	uikit.Base
	text     string
	textFunc func() string
}

func NewLabel(theme *uikit.Theme, text string) *Label {
	cfg := uikit.NewWidgetBaseConfig(theme)
	cfg.DrawSurface = false
	cfg.DrawBorder = false

	base := uikit.NewBase(cfg)

	return &Label{
		Base: base,
		text: text,
	}
}

func (w *Label) Focusable() bool {
	return false
}

func (w *Label) SetText(s string) {
	w.text = s
}

func (w *Label) SetTextFunc(fn func() string) {
	w.textFunc = fn
}

func (w *Label) currentText() string {
	if w.textFunc != nil {
		return w.textFunc()
	}

	return w.text
}

func (w *Label) Update(ctx *uikit.Context) {
	r := w.Measure(false)
	if r.Dy() == 0 {
		w.SetFrame(r.Min.X, r.Min.Y, r.Dx())
	}
}

func (w *Label) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Base.Draw(ctx, dst)

	met, _ := uikit.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	baselineY := r.Min.Y + (r.Dy()-met.Height)/2 + met.Ascent

	ctx.Text.SetColor(ctx.Theme.MutedText)
	ctx.Text.SetAlign(etxt.Left)

	ctx.Text.Draw(dst, w.currentText(), r.Min.X, baselineY)
}

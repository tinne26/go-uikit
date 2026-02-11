package widget

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

var _ uikit.Widget = (*Label)(nil)

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

	ctx.Theme().Text().SetColor(ctx.Theme().MutedTextColor)
	ctx.Theme().Text().SetAlign(etxt.Left | etxt.VertCenter)

	ctx.Theme().Text().Draw(dst, w.currentText(), r.Min.X, r.Min.Y+(r.Dy()/2))
}

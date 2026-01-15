package widget

import (
	"image"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

type Checkbox struct {
	uikit.Base

	label   string
	checked bool
}

func NewCheckbox(theme *uikit.Theme, label string) *Checkbox {
	cfg := uikit.NewWidgetBaseConfig(theme)

	w := &Checkbox{}
	w.label = label

	w.Base = uikit.NewBase(cfg)
	w.Base.On(uikit.EventClick, w.onClick, false)
	return w
}

func (w *Checkbox) Focusable() bool { return true }

func (w *Checkbox) SetChecked(v bool) {
	w.checked = v
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

func (w *Checkbox) Checked() bool { return w.checked }

func (w *Checkbox) onClick(e uikit.Event) bool {
	if !w.IsEnabled() {
		return false
	}

	if e.Type == uikit.EventClick {
		w.SetChecked(!w.Checked())
	}

	return false
}

func (w *Checkbox) Update(ctx *uikit.Context) {
	r := w.Measure(false)
	if r.Dy() == 0 {
		w.SetFrame(r.Min.X, r.Min.Y, r.Dx())
	}

	if !w.IsEnabled() {
		return
	}

	if w.IsFocused() && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		w.SetChecked(!w.Checked())
	}
}

func (w *Checkbox) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	w.Base.Draw(ctx, dst)

	r := w.Measure(false)

	content := common.Inset(r, ctx.Theme.PadX, ctx.Theme.PadY)
	boxSize := ctx.Theme.CheckSize
	if boxSize < 10 {
		boxSize = 10
	}

	boxY := r.Min.Y + (r.Dy()-boxSize)/2
	box := image.Rect(content.Min.X, boxY, content.Min.X+boxSize, boxY+boxSize)

	w.Base.DrawRoundedRect(dst, box, int(float64(boxSize)*0.25), ctx.Theme.Bg)
	w.Base.DrawRoundedBorder(dst, box, int(float64(boxSize)*0.25), ctx.Theme.BorderW, ctx.Theme.Border)

	if w.checked {
		x1 := float32(box.Min.X) + float32(box.Dx())*0.22
		y1 := float32(box.Min.Y) + float32(box.Dy())*0.55
		x2 := float32(box.Min.X) + float32(box.Dx())*0.43
		y2 := float32(box.Min.Y) + float32(box.Dy())*0.73
		x3 := float32(box.Min.X) + float32(box.Dx())*0.78
		y3 := float32(box.Min.Y) + float32(box.Dy())*0.28

		w := float32(ctx.Theme.BorderW)
		if w < 2 {
			w = 2
		}

		vector.StrokeLine(dst, x1, y1, x2, y2, w, ctx.Theme.Focus, false)
		vector.StrokeLine(dst, x2, y2, x3, y3, w, ctx.Theme.Focus, false)
	}

	met, _ := uikit.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	baselineY := r.Min.Y + (r.Dy()-met.Height)/2 + met.Ascent
	tx := box.Max.X + ctx.Theme.SpaceS

	col := ctx.Theme.Text
	if !w.IsEnabled() {
		col = ctx.Theme.Disabled
	}

	ctx.Text.SetColor(col)
	ctx.Text.SetAlign(etxt.Left)
	ctx.Text.Draw(dst, w.label, tx, baselineY)
}

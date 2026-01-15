package widget

import (
	"image"
	"math"

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

	w := &Checkbox{
		label: label,
	}

	w.Base = uikit.NewBase(cfg)

	// Clicking anywhere on the widget triggers toggle (Base must emit EventClick).
	w.Base.On(uikit.EventClick, w.onClick, false)

	return w
}

func (w *Checkbox) Focusable() bool { return true }

func (w *Checkbox) SetChecked(v bool) {
	if w.checked == v {
		return
	}
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

	// Keyboard toggle
	if w.IsFocused() && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		w.SetChecked(!w.Checked())
	}
}

func (w *Checkbox) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	w.Base.Draw(ctx, dst)

	theme := ctx.Theme()
	r := w.Measure(false)

	content := common.Inset(r, theme.PadX, theme.PadY)

	// Checkbox square size
	boxSize := theme.CheckSize
	if boxSize < 12 {
		boxSize = 12
	}

	boxY := r.Min.Y + (r.Dy()-boxSize)/2
	box := image.Rect(content.Min.X, boxY, content.Min.X+boxSize, boxY+boxSize)

	// Colors
	bg := theme.BackgroundColor
	border := theme.BorderColor
	checkCol := theme.FocusColor
	textCol := theme.TextColor

	if !w.IsEnabled() {
		border = theme.DisabledColor
		checkCol = theme.DisabledColor
		textCol = theme.DisabledColor
	}

	radius := int(math.Round(float64(boxSize) * 0.22))
	if radius < 2 {
		radius = 2
	}

	if w.IsEnabled() {
		if w.IsPressed() {
			bg = theme.FocusColor
		} else if w.IsHovered() {
			bg = theme.BorderColor
		}
	}

	w.Base.DrawRoundedRect(dst, box, radius, bg)
	w.Base.DrawRoundedBorder(dst, box, radius, theme.BorderW, border)

	if w.checked {
		x1 := float32(box.Min.X) + float32(boxSize)*0.22
		y1 := float32(box.Min.Y) + float32(boxSize)*0.55
		x2 := float32(box.Min.X) + float32(boxSize)*0.42
		y2 := float32(box.Min.Y) + float32(boxSize)*0.73
		x3 := float32(box.Min.X) + float32(boxSize)*0.78
		y3 := float32(box.Min.Y) + float32(boxSize)*0.30

		strokeW := float32(theme.BorderW)
		if strokeW < 2 {
			strokeW = 2
		}
		// Slightly thicker for better readability at small sizes
		if strokeW < float32(boxSize)/8 {
			strokeW = float32(boxSize) / 8
		}

		vector.StrokeLine(dst, x1, y1, x2, y2, strokeW, checkCol, true)
		vector.StrokeLine(dst, x2, y2, x3, y3, strokeW, checkCol, true)
	}

	tx := box.Max.X + theme.SpaceS

	t := theme.Text()
	t.SetColor(textCol)
	t.SetAlign(etxt.Left | etxt.VertCenter)
	t.Draw(dst, w.label, tx, r.Min.Y+r.Dy()/2)
}

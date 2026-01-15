package widget

import (
	"math"
	"time"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

// TextInput is a single-line input box (no label).
// Height and proportions come from Theme; external layout controls only width.
type TextInput struct {
	uikit.Base

	text        string
	placeholder string
	caretTick   int
}

func NewTextInput(theme *uikit.Theme, placeholder string) *TextInput {
	cfg := uikit.NewWidgetBaseConfig(theme)

	w := &TextInput{
		placeholder: placeholder,
	}

	w.Base = uikit.NewBase(cfg)
	return w
}

func (w *TextInput) Focusable() bool {
	return true
}

func (w *TextInput) WantsIME() bool {
	return true
}

func (w *TextInput) Text() string {
	return w.text
}

func (w *TextInput) SetText(s string) {
	w.text = s
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

func (w *TextInput) AppendText(s string) {
	w.SetText(w.Text() + s)
}

func (w *TextInput) Reset() {
	w.SetText("")
}

func (w *TextInput) Update(ctx *uikit.Context) {
	r := w.Measure(false)
	if r.Dy() == 0 {
		w.SetFrame(r.Min.X, r.Min.Y, r.Max.X)
	}

	if w.IsFocused() && w.IsEnabled() {
		w.caretTick++
	} else {
		w.caretTick = 0
	}

	if !w.IsFocused() || !w.IsEnabled() {
		return
	}

	for _, r := range ebiten.AppendInputChars(nil) {
		if r == '\b' || r == 0x7f {
			w.backspace()
			continue
		}

		if r < 0x20 {
			continue
		}

		w.AppendText(string(r))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		w.backspace()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		w.backspace()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		ctx.SetFocus(nil)
	}
}

func (w *TextInput) backspace() {
	if w.text == "" {
		return
	}

	rs := []rune(w.text)
	if len(rs) == 0 {
		w.SetText("")
		return
	}

	w.SetText(string(rs[:len(rs)-1]))
}

func (w *TextInput) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	w.Base.Draw(ctx, dst)
	r := w.Measure(false)

	content := common.Inset(r, ctx.Theme.PadX, ctx.Theme.PadY)

	met, _ := uikit.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	baselineY := r.Min.Y + (r.Dy()-met.Height)/2 + met.Ascent

	drawStr := w.text
	textCol := ctx.Theme.Text
	if drawStr == "" && !w.IsFocused() {
		drawStr = w.placeholder
		textCol = ctx.Theme.MutedText
	}

	textW := uikit.MeasureStringPx(ctx.Theme.Font, ctx.Theme.FontPx, drawStr)
	shiftX := 0
	if textW > content.Dx() {
		shiftX = content.Dx() - textW
	}

	ctx.Text.SetAlign(etxt.Left)
	ctx.Text.SetColor(textCol)
	ctx.Text.Draw(dst, drawStr, content.Min.X+shiftX, baselineY)

	if w.IsFocused() && w.IsEnabled() && w.Theme().CaretWidthPx > 0 {
		blinkFrames := int(math.Max(1, float64(w.Theme().CaretBlink)/float64(time.Second)*60.0))

		if (w.caretTick/blinkFrames)%2 == 0 {
			wBefore := uikit.MeasureStringPx(ctx.Theme.Font, ctx.Theme.FontPx, w.text)
			cx := content.Min.X + shiftX + wBefore + w.Theme().CaretMarginPx
			cy := baselineY - met.Ascent
			caretH := met.Height

			if cx < content.Min.X {
				cx = content.Min.X
			}

			if cx > content.Min.X+content.Dx() {
				cx = content.Min.X + content.Dx()
			}

			vector.DrawFilledRect(dst, float32(cx), float32(cy), float32(w.Theme().CaretWidthPx), float32(caretH), ctx.Theme.Caret, false)
		}
	}
}

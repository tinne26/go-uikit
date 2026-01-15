package widget

import (
	"math"
	"strings"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/common"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tinne26/etxt"
)

// TextArea is a multi-line text editor with internal vertical scrolling.
type TextArea struct {
	uikit.Base

	text        string
	placeholder string

	lines int

	Scroll uikit.Scroller

	// Caret config (end-of-text caret)
	CaretWidthPx  int
	CaretBlinkMs  int
	CaretMarginPx int
	caretTick     int
}

func NewTextArea(theme *uikit.Theme, placeholder string) *TextArea {
	cfg := uikit.NewWidgetBaseConfig(theme)

	t := &TextArea{
		Base:          uikit.NewBase(cfg),
		placeholder:   placeholder,
		lines:         5,
		CaretWidthPx:  2,
		CaretBlinkMs:  600,
		CaretMarginPx: 0,
	}

	t.Scroll = uikit.NewScroller()
	t.Scroll.Scrollbar = uikit.ScrollbarAlways // textarea typically shows it (content-dependent)

	t.Base.HeightCaculator = t.calculateHeight
	return t
}

func (w *TextArea) calculateHeight() int {
	met, _ := uikit.MetricsPx(w.Theme().Font, w.Theme().FontPx)
	lines := w.lines
	if lines <= 0 {
		lines = 5
	}

	controlH := w.Theme().PadY*2 + lines*met.Height
	if controlH < w.Theme().ControlH {
		controlH = w.Theme().ControlH
	}

	return controlH
}

func (w *TextArea) Focusable() bool {
	return true
}

func (w *TextArea) WantsIME() bool {
	return true
}

func (w *TextArea) Text() string {
	return w.text
}

func (w *TextArea) SetText(s string) {
	w.text = s
	w.Dispatch(uikit.Event{Widget: w, Type: uikit.EventValueChange})
}

func (w *TextArea) AppendText(s string) {
	w.SetText(w.Text() + s)
}

func (w *TextArea) SetLines(n int) {
	if n < 1 {
		n = 1
	}

	w.lines = n
}

func (w *TextArea) Update(ctx *uikit.Context) {
	r := w.Measure(false)
	if r.Dx() > 0 && r.Dy() == 0 {
		w.SetFrame(r.Min.X, r.Min.Y, r.Dx())
	}

	if w.IsFocused() && w.IsEnabled() {
		w.caretTick++
	} else {
		w.caretTick = 0
	}

	content := common.Inset(r, ctx.Theme.PadX, ctx.Theme.PadY)

	met, _ := uikit.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)

	lineCount := 1
	if w.text != "" {
		lineCount = 1 + strings.Count(w.text, "\n")
	}

	contentH := lineCount * met.Height
	if contentH < content.Dy() {
		contentH = content.Dy()
	}

	w.Scroll.Update(ctx, content, contentH)

	if !w.IsFocused() || !w.IsEnabled() {
		return
	}

	changed := false
	for _, r := range ebiten.AppendInputChars(nil) {
		if r == '\b' || r == 0x7f {
			w.backspace()
			changed = true
			continue
		}

		if r == '\n' || r == '\r' {
			w.AppendText("\n")
			changed = true
			continue
		}

		if r < 0x20 {
			continue
		}

		w.AppendText(string(r))
		changed = true
	}

	// Optional: keep these for platforms that don't send backspace/enter via chars
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		w.backspace()
		changed = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		w.AppendText("\n")
		changed = true
	}

	// Scroll-to-caret (caret at end of text)
	if changed {
		parts := strings.Split(w.text, "\n")
		lastIdx := len(parts) - 1
		caretBottom := (lastIdx + 1) * met.Height

		minScroll := caretBottom - content.Dy()
		if minScroll < 0 {
			minScroll = 0
		}
		if w.Scroll.ScrollY < minScroll {
			w.Scroll.ScrollY = minScroll
		}

		maxScroll := contentH - content.Dy()
		if maxScroll < 0 {
			maxScroll = 0
		}

		if w.Scroll.ScrollY > maxScroll {
			w.Scroll.ScrollY = maxScroll
		}
	}
}

func (w *TextArea) backspace() {
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

func (w *TextArea) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	r := w.Measure(false)
	if r.Dx() > 0 && r.Dy() == 0 {
		w.SetFrame(r.Min.X, r.Min.Y, r.Dx())
	}

	content := common.Inset(r, ctx.Theme.PadX, ctx.Theme.PadY)

	w.DrawSurfece(ctx, dst, r)
	w.DrawBoder(ctx, dst, r)
	w.DrawFocus(ctx, dst, r)

	// Clip to content area
	sub := dst.SubImage(content).(*ebiten.Image)
	ox, oy := sub.Bounds().Min.X, sub.Bounds().Min.Y

	// Ensure renderer state for content
	ctx.Text.SetFont(ctx.Theme.Font)
	ctx.Text.SetSize(float64(ctx.Theme.FontPx))
	ctx.Text.SetAlign(etxt.Left)

	met, _ := uikit.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	startY := -w.Scroll.ScrollY

	// Placeholder
	drawStr := w.text
	col := ctx.Theme.Text
	if drawStr == "" && !w.IsFocused() {
		drawStr = w.placeholder
		col = ctx.Theme.MutedText
	}
	ctx.Text.SetColor(col)

	lines := strings.Split(drawStr, "\n")
	for i, line := range lines {
		y := startY + i*met.Height + met.Ascent
		ctx.Text.Draw(sub, line, ox, oy+y)
	}

	// Scrollbar (SubImage-safe after patch below)
	lineCount := 1
	if w.text != "" {
		lineCount = 1 + strings.Count(w.text, "\n")
	}
	contentH := lineCount * met.Height
	if contentH < content.Dy() {
		contentH = content.Dy()
	}

	w.Scroll.DrawBar(sub, ctx.Theme, content.Dx(), content.Dy(), contentH)

	// Caret at end (approx). Draw on sub so it's clipped.
	if w.IsFocused() && w.IsEnabled() && w.CaretWidthPx > 0 {
		blinkFrames := int(math.Max(1, float64(w.CaretBlinkMs)/1000.0*60.0))
		if (w.caretTick/blinkFrames)%2 == 0 {
			lastLine := ""
			lastIdx := 0
			if w.text != "" {
				parts := strings.Split(w.text, "\n")
				lastIdx = len(parts) - 1
				lastLine = parts[lastIdx]
			}

			wBefore := uikit.MeasureStringPx(ctx.Theme.Font, ctx.Theme.FontPx, lastLine)
			cx := wBefore + w.CaretMarginPx
			cy := (lastIdx * met.Height) - w.Scroll.ScrollY

			if cx < 0 {
				cx = 0
			}
			maxX := content.Dx() - w.CaretWidthPx
			if maxX < 0 {
				maxX = 0
			}
			if cx > maxX {
				cx = maxX
			}

			if cy+met.Height > 0 && cy <= content.Dy() {
				// IMPORTANT: SubImage keeps absolute coords -> add (ox,oy)
				vector.DrawFilledRect(
					sub,
					float32(ox+cx),
					float32(oy+cy),
					float32(w.CaretWidthPx),
					float32(met.Height),
					ctx.Theme.Caret,
					false,
				)
			}
		}
	}

	w.DrawInvalid(ctx, dst, r)
}

package ui

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TextArea is a multi-line text editor with internal vertical scrolling.
// Notes:
// - It supports new lines (Enter).
// - Editing is simple append-at-end (no caret navigation yet).
// - Height can be configured in "visible lines" via SetLines().
type TextArea struct {
	base  Base
	theme *Theme

	text        string
	placeholder string

	// Visible line count (controls height). If <= 0, defaults to 5.
	lines int

	// Vertical scroll in pixels (content space)
	scrollY int

	// Drag scroll
	dragging bool
	lastPY   int

	// Caret config (end-of-text caret)
	CaretWidthPx  int
	CaretBlinkMs  int
	CaretMarginPx int
	caretTick     int
}

func NewTextArea(placeholder string) *TextArea {
	return &TextArea{
		base:          NewBase(),
		placeholder:   placeholder,
		lines:         5,
		CaretWidthPx:  2,
		CaretBlinkMs:  600,
		CaretMarginPx: 0,
	}
}

func (t *TextArea) Base() *Base     { return &t.base }
func (t *TextArea) Focusable() bool { return true }
func (t *TextArea) WantsIME() bool  { return true }

func (t *TextArea) Text() string     { return t.text }
func (t *TextArea) SetText(s string) { t.text = s }

// SetLines sets how many text lines are visible (controls the widget height).
func (t *TextArea) SetLines(n int) {
	if n < 1 {
		n = 1
	}
	t.lines = n
}

func (t *TextArea) SetRectByWidth(x, y, w int) {
	if t.theme == nil {
		t.base.Rect = Rect{X: x, Y: y, W: w, H: 0}
		return
	}

	met, _ := MetricsPx(t.theme.Font, t.theme.FontPx)

	lines := t.lines
	if lines <= 0 {
		lines = 5
	}

	// Control height = vertical padding + N lines of text height.
	controlH := t.theme.PadY*2 + lines*met.Height
	// Ensure it's not smaller than the standard control height.
	if controlH < t.theme.ControlH {
		controlH = t.theme.ControlH
	}

	h := controlH

	// Reserve space for validation message (same behavior as Base.RequiredHeight).
	if t.base.Invalid && t.base.ErrorText != "" {
		em, _ := MetricsPx(t.theme.Font, t.theme.ErrorFontPx)
		h += t.theme.ErrorGap + em.Height
	}

	t.base.Rect = Rect{X: x, Y: y, W: w, H: h}
}

func (t *TextArea) Update(ctx *Context) {
	t.theme = ctx.Theme
	if t.base.Rect.H == 0 {
		t.SetRectByWidth(t.base.Rect.X, t.base.Rect.Y, t.base.Rect.W)
	}

	if t.base.focused && t.base.Enabled {
		t.caretTick++
	} else {
		t.caretTick = 0
	}

	ctrl := t.base.ControlRect(ctx.Theme)
	content := ctrl.Inset(ctx.Theme.PadX, ctx.Theme.PadY)

	// Wheel scroll (desktop)
	_, wy := ebiten.Wheel()
	if wy != 0 && content.Contains(ctx.ptrX, ctx.ptrY) {
		step := int(math.Round(float64(content.H) * 0.20))
		if step < 8 {
			step = 8
		}
		t.scrollY -= int(math.Round(wy * float64(step)))
	}

	// Drag scroll (touch or mouse)
	if ctx.ptrJustDown && content.Contains(ctx.ptrX, ctx.ptrY) {
		t.dragging = true
		t.lastPY = ctx.ptrY
	}
	if t.dragging && ctx.ptrDown {
		dy := ctx.ptrY - t.lastPY
		t.scrollY -= dy
		t.lastPY = ctx.ptrY
	}
	if ctx.ptrJustUp {
		t.dragging = false
	}

	t.clampScroll(ctx, content)

	// Editing only when focused
	if !t.base.focused || !t.base.Enabled {
		return
	}

	for _, r := range ebiten.AppendInputChars(nil) {
		if r == '\b' || r == 0x7f {
			t.backspace()
			continue
		}
		if r == '\n' || r == '\r' {
			t.text += "\n"
			continue
		}
		if r < 0x20 {
			continue
		}
		t.text += string(r)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		t.backspace()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		t.text += "\n"
	}
}

func (t *TextArea) backspace() {
	if t.text == "" {
		return
	}
	rs := []rune(t.text)
	if len(rs) == 0 {
		t.text = ""
		return
	}
	t.text = string(rs[:len(rs)-1])
}

func (t *TextArea) clampScroll(ctx *Context, content Rect) {
	met, _ := MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	lines := 1
	if t.text != "" {
		lines = 1 + strings.Count(t.text, "\n")
	}
	totalH := lines * met.Height
	maxScroll := totalH - content.H
	if maxScroll < 0 {
		maxScroll = 0
	}
	if t.scrollY < 0 {
		t.scrollY = 0
	}
	if t.scrollY > maxScroll {
		t.scrollY = maxScroll
	}
}

func (t *TextArea) Draw(ctx *Context, dst *ebiten.Image) {
	t.theme = ctx.Theme
	if t.base.Rect.H == 0 {
		t.SetRectByWidth(t.base.Rect.X, t.base.Rect.Y, t.base.Rect.W)
	}

	r := t.base.ControlRect(ctx.Theme)

	// Surface
	bg := ctx.Theme.Surface
	if !t.base.Enabled {
		bg = ctx.Theme.SurfacePressed
	} else if t.base.pressed {
		bg = ctx.Theme.SurfacePressed
	} else if t.base.hovered {
		bg = ctx.Theme.SurfaceHover
	}
	drawRoundedRect(dst, r, ctx.Theme.Radius, bg)

	// Border (red when invalid)
	borderCol := ctx.Theme.Border
	if t.base.Invalid {
		borderCol = ctx.Theme.ErrorBorder
	}
	drawRoundedBorder(dst, r, ctx.Theme.Radius, ctx.Theme.BorderW, borderCol)

	// Focus ring
	if t.base.focused && t.base.Enabled {
		drawFocusRing(dst, r, ctx.Theme.Radius, ctx.Theme.FocusRingGap, ctx.Theme.FocusRingW, ctx.Theme.Focus)
	}

	content := r.Inset(ctx.Theme.PadX, ctx.Theme.PadY)
	sub := dst.SubImage(content.ImageRect()).(*ebiten.Image)

	met, _ := MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	startY := -t.scrollY

	drawStr := t.text
	col := ctx.Theme.Text
	if drawStr == "" && !t.base.focused {
		drawStr = t.placeholder
		col = ctx.Theme.MutedText
	}

	ctx.Text.SetAlign(0)
	ctx.Text.SetColor(col)

	// Render line by line
	lines := strings.Split(drawStr, "\n")
	for i, line := range lines {
		y := startY + i*met.Height + met.Ascent
		ctx.Text.Draw(sub, line, 0, y)
	}

	// Scrollbar (if needed)
	totalLines := 1
	if t.text != "" {
		totalLines = 1 + strings.Count(t.text, "\n")
	}
	totalH := totalLines * met.Height
	if totalH > content.H {
		trackW := int(math.Max(3, float64(ctx.Theme.BorderW)))
		trackX := content.Right() - trackW
		trackY := content.Y
		trackH := content.H

		thumbH := int(math.Max(12, float64(trackH)*float64(trackH)/float64(totalH)))
		maxScroll := totalH - trackH
		thumbY := trackY
		if maxScroll > 0 {
			thumbY = trackY + int(math.Round(float64(trackH-thumbH)*float64(t.scrollY)/float64(maxScroll)))
		}
		vector.DrawFilledRect(dst, float32(trackX), float32(trackY), float32(trackW), float32(trackH), ctx.Theme.Border, false)
		vector.DrawFilledRect(dst, float32(trackX), float32(thumbY), float32(trackW), float32(thumbH), ctx.Theme.Focus, false)
	}

	// Caret at end (approx: last line end)
	if t.base.focused && t.base.Enabled && t.CaretWidthPx > 0 {
		blinkFrames := int(math.Max(1, float64(t.CaretBlinkMs)/1000.0*60.0))
		if (t.caretTick/blinkFrames)%2 == 0 {
			lastLine := ""
			lastIdx := 0
			if t.text != "" {
				parts := strings.Split(t.text, "\n")
				lastIdx = len(parts) - 1
				lastLine = parts[lastIdx]
			}
			wBefore := MeasureStringPx(ctx.Theme.Font, ctx.Theme.FontPx, lastLine)
			cx := content.X + wBefore + t.CaretMarginPx
			cy := content.Y + (lastIdx*met.Height - t.scrollY)
			caretH := met.Height

			if cx < content.X {
				cx = content.X
			}
			if cx > content.Right() {
				cx = content.Right()
			}
			vector.DrawFilledRect(dst, float32(cx), float32(cy), float32(t.CaretWidthPx), float32(caretH), ctx.Theme.Caret, false)
		}
	}

	// Validation message
	err := t.base.ErrorRect(ctx.Theme)
	if t.base.Invalid {
		drawErrorText(ctx, dst, err, t.base.ErrorText)
	}
}

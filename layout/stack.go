package layout

import (
	"image/color"

	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

// Stack places children vertically. If height > 0 it becomes scrollable and clips via SubImage.
type Stack struct {
	uikit.Base
	children []uikit.Widget

	padX int
	padY int
	gap  int

	Scroll uikit.Scroller

	height   int
	contentH int

	scratch    *ebiten.Image
	background color.RGBA
}

func NewStack(theme *uikit.Theme) *Stack {
	l := &Stack{}

	cfg := uikit.NewWidgetBaseConfig(theme)
	l.Base = uikit.NewBase(cfg)
	l.Base.SetEnabled(true)
	l.Base.HeightCalculator = func() int {
		if l.height == 0 {
			return l.contentH
		}

		return l.height
	}

	l.gap = theme.SpaceS
	l.Scroll = uikit.NewScroller()
	return l
}

func (l *Stack) Focusable() bool { return false }

// SetHeight sets the viewport height. Use 0 for unlimited (no scroll, no clipping).
func (l *Stack) SetHeight(h int) {
	l.height = h
}

func (l *Stack) SetPadding(x, y int) {
	l.padX = x
	l.padY = y
}

func (l *Stack) SetGap(v int) {
	l.gap = v
}

func (l *Stack) Children() []uikit.Widget {
	return l.children
}
func (l *Stack) SetChildren(ws []uikit.Widget) {
	l.children = ws
}
func (l *Stack) Add(ws ...uikit.Widget) {
	l.children = append(l.children, ws...)
}

func (l *Stack) Clear() {
	l.children = nil
}

func (l *Stack) Update(ctx *uikit.Context) {
	l.doLayout(ctx)

	r := l.Measure(false)

	// Scroll input only when height is limited
	if r.Dy() > 0 {
		l.Scroll.Update(ctx, r, l.contentH)
		l.doLayout(ctx)
	}

	for _, w := range l.children {
		if !w.IsVisible() {
			continue
		}

		w.Update(ctx)
	}
}

func (l *Stack) doLayout(ctx *uikit.Context) {
	vp := l.Measure(false)
	x0 := vp.Min.X + l.padX
	y0 := vp.Min.Y + l.padY
	w0 := vp.Dx() - l.padX*2
	if w0 < 0 {
		w0 = 0
	}

	y := y0
	if vp.Dy() > 0 {
		y -= l.Scroll.ScrollY
	}

	contentH := l.padY * 2
	for i, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}

		ch.SetFrame(x0, y, w0)
		r := ch.Measure(true)
		contentH += r.Dy()
		if i != len(l.children)-1 {
			contentH += l.gap
		}
		y += r.Dy() + l.gap
	}

	// At least viewport height so scrollbar math is stable
	if vp.Dy() > 0 && contentH < vp.Dy() {
		contentH = vp.Dy()
	}

	l.contentH = contentH
}

func (l *Stack) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	vp := l.Measure(false)

	if vp.Dy() <= 0 {
		for _, ch := range l.children {
			if !ch.IsVisible() {
				continue
			}

			ch.Draw(ctx, dst)
		}

		return
	}

	// Scrollable: render to a full-screen scratch (no coordinate shifting),
	// then copy only the viewport region back to dst using SubImage.
	sw, sh := dst.Bounds().Dx(), dst.Bounds().Dy()
	if l.scratch == nil || l.scratch.Bounds().Dx() != sw || l.scratch.Bounds().Dy() != sh {
		l.scratch = ebiten.NewImage(sw, sh)
	}

	l.scratch.Clear()

	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}

		ch.Draw(ctx, l.scratch)
	}

	part := l.scratch.SubImage(vp).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(vp.Min.X), float64(vp.Min.Y))
	dst.DrawImage(part, op)

	sub := dst.SubImage(vp).(*ebiten.Image)

	l.Scroll.DrawBar(sub, ctx.Theme(), vp.Dx(), vp.Dy(), l.contentH)
}

func (l *Stack) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	// Overlay should escape clipping -> draw on dst (not on subimage)
	for _, ch := range l.children {
		if ow, ok := any(ch).(uikit.OverlayWidget); ok && ow.OverlayActive() {
			ow.DrawOverlay(ctx, dst)
		}
		// Nested layouts will propagate overlays naturally.
		if ll, ok := any(ch).(interface {
			DrawOverlay(*uikit.Context, *ebiten.Image)
		}); ok {
			ll.DrawOverlay(ctx, dst)
		}
	}
}

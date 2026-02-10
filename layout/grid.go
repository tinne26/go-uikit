package layout

import (
	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

var _ uikit.Layout = (*Grid)(nil)

// Grid places children in a fixed column grid. If height > 0 it becomes scrollable and clips via SubImage.
type Grid struct {
	uikit.Base
	children []uikit.Widget
	scroll   uikit.Scroller

	columns int
	padX    int
	padY    int
	gapX    int
	gapY    int
	height  int
	scratch *ebiten.Image
}

func NewGrid(theme *uikit.Theme) *Grid {
	l := &Grid{}
	l.columns = 2
	l.gapX = theme.SpaceS
	l.gapY = theme.SpaceS
	l.scroll = uikit.NewScroller()

	cfg := uikit.NewWidgetBaseConfig(theme)
	l.Base = uikit.NewBase(cfg)
	l.Base.HeightCalculator = func() int {
		return l.height
	}

	return l
}

func (l *Grid) Focusable() bool { return false }

func (l *Grid) SetHeight(h int) {
	l.height = h
}

func (l *Grid) SetPadding(x, y int) {
	l.padX = x
	l.padY = y
}

func (l *Grid) SetGap(x, y int) {
	l.gapX = x
	l.gapY = y
}

func (l *Grid) SetColumns(c int) {
	l.columns = c
}

func (l *Grid) Children() []uikit.Widget {
	return l.children
}

func (l *Grid) SetChildren(ws []uikit.Widget) {
	l.children = ws
}

func (l *Grid) Add(ws ...uikit.Widget) {
	l.children = append(l.children, ws...)
}

func (l *Grid) Clear() {
	l.children = nil
}

func (l *Grid) Update(ctx *uikit.Context) {
	l.doLayout(ctx)

	if l.Measure(false).Dy() > 0 {
		l.scroll.Update(ctx, l.Measure(false), l.height)
		l.doLayout(ctx)
	}

	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.Update(ctx)
	}
}

func (l *Grid) doLayout(ctx *uikit.Context) {
	vp := l.Measure(false)
	cols := l.columns
	if cols <= 0 {
		cols = 2
	}

	innerW := vp.Dx() - l.padX*2
	if innerW < 0 {
		innerW = 0
	}
	cellW := innerW
	if cols > 0 {
		cellW = (innerW - (cols-1)*l.gapX) / cols
		if cellW < 0 {
			cellW = 0
		}
	}

	x0 := vp.Min.X + l.padX
	y0 := vp.Min.Y + l.padY
	x := x0
	y := y0
	if vp.Dy() > 0 {
		y -= l.scroll.ScrollY
	}

	contentH := l.padY * 2
	rowMaxH := 0
	col := 0

	for i, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.SetFrame(x, y, cellW)
		r := ch.Measure(false)
		if r.Dy() > rowMaxH {
			rowMaxH = r.Dy()
		}

		col++
		last := i == len(l.children)-1
		if col >= cols || last {
			contentH += rowMaxH
			if !last {
				contentH += l.gapY
			}
			y += rowMaxH + l.gapY
			x = x0
			col = 0
			rowMaxH = 0
		} else {
			x += cellW + l.gapX
		}
	}

	if vp.Dy() > 0 && contentH < vp.Dy() {
		contentH = vp.Dy()
	}

	l.SetHeight(contentH)
}

func (l *Grid) Draw(ctx *uikit.Context, dst *ebiten.Image) {
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
	l.scroll.DrawBar(sub, ctx.Theme(), vp.Dx(), vp.Dy(), l.height)
}

func (l *Grid) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	for _, ch := range l.children {
		if ow, ok := any(ch).(uikit.OverlayWidget); ok && ow.OverlayActive() {
			ow.DrawOverlay(ctx, dst)
		}
		if ll, ok := any(ch).(interface {
			DrawOverlay(*uikit.Context, *ebiten.Image)
		}); ok {
			ll.DrawOverlay(ctx, dst)
		}
	}
}

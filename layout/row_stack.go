package layout

import (
	"math"
	"slices"

	"github.com/erparts/go-uikit"
	"github.com/hajimehoshi/ebiten/v2"
)

type rowStackPattern struct {
	Weights []float64
	Gap     int
}

func (rsp *rowStackPattern) normalize() {
	var sum float64
	for i, weight := range rsp.Weights {
		if weight < 0 {
			rsp.Weights[i] = 0
		} else {
			sum += weight
		}
	}

	if sum > 0 {
		for i, weight := range rsp.Weights {
			rsp.Weights[i] = weight / sum
		}
	} else {
		div := float64(len(rsp.Weights))
		for i := range rsp.Weights {
			rsp.Weights[i] = 1.0 / div
		}
	}
}

func (rsp *rowStackPattern) computeWidths(totalWidth int, widthsBuffer []int) []int {
	widthsBuffer = widthsBuffer[:0]

	//fmt.Printf("computeWidths: totalWidth = %d\n", totalWidth)
	totalWidth -= rsp.Gap * max(len(rsp.Weights)-1, 0)
	if totalWidth <= 0 {
		for range len(rsp.Weights) {
			widthsBuffer = append(widthsBuffer, 0)
		}
		return widthsBuffer
	}

	tw := float64(totalWidth)
	accWidth := 0
	accWeight := 0.0
	lastNonZero := -1
	for i, weight := range rsp.Weights {
		if weight == 0 {
			widthsBuffer = append(widthsBuffer, 0)
			continue
		}

		accWeight += weight
		expWidth := int(math.Round(accWeight * tw))
		widthInt := expWidth - accWidth
		accWidth += widthInt
		if widthInt > 0 || lastNonZero == -1 {
			lastNonZero = i
		}
		widthsBuffer = append(widthsBuffer, widthInt)
	}

	if accWidth > totalWidth {
		widthsBuffer[lastNonZero] -= 1
	} else if accWidth < totalWidth {
		widthsBuffer[lastNonZero] += 1
	}

	return widthsBuffer
}

var _ uikit.Layout = (*RowStack)(nil)

// RowStack is a top to bottom layout where each row can have a
// different number of widgets and a different amount of space
// for each.
type RowStack struct {
	uikit.Base
	uikit.Scroller

	patterns   map[int]rowStackPattern
	children   []uikit.Widget
	padX, padY int
	height     int
	contentH   int
	rowGap     int

	widthsBuffer []int
}

// NewRowStack creates a RowStack with the given default pattern.
func NewRowStack(theme *uikit.Theme, widthWeights ...float64) *RowStack {
	defaultPattern := rowStackPattern{Weights: widthWeights, Gap: theme.SpaceS}
	defaultPattern.normalize()

	cfg := uikit.NewWidgetBaseConfig(theme)
	l := &RowStack{
		Base:     uikit.NewBase(cfg),
		Scroller: uikit.NewScroller(),
		patterns: map[int]rowStackPattern{-1: defaultPattern},
	}

	l.Base.HeightCalculator = l.heightCalculator
	return l
}

func (l *RowStack) heightCalculator() int {
	if l.height == 0 {
		return l.contentH
	}
	return l.height
}

// DefaultPattern returns the properties of the default row pattern.
func (l *RowStack) DefaultPattern() (gap int, normWidthWeights []float64) {
	pattern := l.patterns[-1]
	return pattern.Gap, pattern.Weights
}

// SetRowPattern sets the widget distribution pattern of a specific row.
// If rowIndex < 0, the given pattern is set as the default.
func (l *RowStack) SetRowPattern(rowIndex int, itemGap int, widthWeights ...float64) {
	pattern := rowStackPattern{Weights: slices.Clone(widthWeights), Gap: itemGap}
	pattern.normalize()
	rowIndex = max(rowIndex, -1)
	l.patterns[rowIndex] = pattern
}

func (l *RowStack) SetRowGap(gap int) {
	l.rowGap = gap
}

func (l *RowStack) Focusable() bool { return false }

func (l *RowStack) SetHeight(h int) {
	l.height = h
}

func (l *RowStack) SetPadding(x, y int) {
	l.padX = x
	l.padY = y
}

func (l *RowStack) Children() []uikit.Widget {
	return l.children
}

func (l *RowStack) SetChildren(ws []uikit.Widget) {
	l.children = ws
}

func (l *RowStack) Add(ws ...uikit.Widget) {
	l.children = append(l.children, ws...)
}

func (l *RowStack) Clear() {
	l.children = l.children[:0]
}

func (l *RowStack) Update(ctx *uikit.Context) {
	if l.height > 0 {
		l.Scroller.Update(ctx, l.Measure(false), l.height)
	}
	l.doLayout(ctx)

	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.Update(ctx)
	}
}

func (l *RowStack) doLayout(ctx *uikit.Context) {
	area := l.Measure(false)

	ox := area.Min.X + l.padX
	y := area.Min.Y + l.padY - l.Scroller.ScrollY
	contentWidth := max(area.Dx()-l.padX*2, 0)

	basePattern := l.patterns[-1]

	l.contentH = l.padY * 2
	var anyVisible bool
	rowIndex := 0
	childIndex := 0
	for {
		pattern, ok := l.patterns[rowIndex]
		if !ok {
			pattern = basePattern
		}

		l.widthsBuffer = pattern.computeWidths(contentWidth, l.widthsBuffer)

		maxHeight := 0
		colIndex := 0
		x := ox
		for _, child := range l.children[childIndex:] {
			if !child.IsVisible() {
				continue
			}

			anyVisible = true
			width := l.widthsBuffer[colIndex]
			child.SetFrame(x, y, width)
			x += width + pattern.Gap
			colIndex += 1
		}
		if colIndex == 0 {
			break
		}
		l.contentH += maxHeight + l.rowGap
		y += maxHeight + l.rowGap

		rowIndex += 1
	}

	if anyVisible {
		l.contentH -= l.rowGap
	}
}

func (l *RowStack) Draw(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	r := l.Measure(false)
	sub := dst.SubImage(r).(*ebiten.Image)
	for _, ch := range l.children {
		if !ch.IsVisible() {
			continue
		}
		ch.Draw(ctx, sub)
	}
	if l.height > 0 {
		l.Scroller.DrawBar(sub, ctx.Theme(), sub.Bounds().Dx(), sub.Bounds().Dy(), l.contentH)
	}
}

type OverlayDrawer interface {
	DrawOverlay(*uikit.Context, *ebiten.Image)
}

func (l *RowStack) DrawOverlay(ctx *uikit.Context, dst *ebiten.Image) {
	if !l.IsVisible() {
		return
	}

	for _, child := range l.children {
		switch overlay := child.(type) {
		case uikit.OverlayWidget:
			if overlay.OverlayActive() {
				overlay.DrawOverlay(ctx, dst)
			}
		case OverlayDrawer: // TODO: maybe simplify to only uikit.OverlayWidget and adjust widget.Select?
			overlay.DrawOverlay(ctx, dst)
		}
	}
}

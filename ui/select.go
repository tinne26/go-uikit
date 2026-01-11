package ui

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Select is a simple dropdown selector.
// When open, it expands downward and pushes layout (caller must SetRectByWidth each frame).
type Select struct {
	base  Base
	theme *Theme

	options []string
	index   int

	open bool

	// internal
	scroll int

	// MaxVisible controls how many options are shown when open.
	MaxVisible int
}

func NewSelect(options []string) *Select {
	return &Select{
		base:       NewBase(),
		options:    options,
		index:      0,
		MaxVisible: 5,
	}
}

func (s *Select) Base() *Base     { return &s.base }
func (s *Select) Focusable() bool { return true }

func (s *Select) SetOptions(opts []string) {
	s.options = opts
	if s.index >= len(opts) {
		s.index = 0
	}
}

func (s *Select) Index() int { return s.index }
func (s *Select) Value() string {
	if s.index < 0 || s.index >= len(s.options) {
		return ""
	}
	return s.options[s.index]
}

func (s *Select) SetIndex(i int) {
	if i < 0 {
		i = 0
	}
	if i >= len(s.options) {
		i = len(s.options) - 1
	}
	if len(s.options) == 0 {
		i = 0
	}
	s.index = i
}

func (s *Select) SetRectByWidth(x, y, w int) {
	if s.theme != nil {
		// Height depends on open state + invalid.
		h := s.base.RequiredHeight(s.theme)
		if s.open {
			n := len(s.options)
			max := s.MaxVisible
			if max <= 0 {
				max = 5
			}
			if n > max {
				n = max
			}
			h += s.theme.SpaceS + n*s.theme.ControlH
		}
		s.base.Rect = Rect{X: x, Y: y, W: w, H: h}
		return
	}
	s.base.Rect = Rect{X: x, Y: y, W: w, H: 0}
}

func (s *Select) HandleEvent(ctx *Context, e Event) {
	if !s.base.Enabled {
		return
	}
	if e.Type == EventClick {
		// Toggle open when clicking on the control area
		ctrl := s.base.ControlRect(ctx.Theme)
		if ctrl.Contains(e.X, e.Y) {
			s.open = !s.open
		} else if s.open {
			// Selecting options
			listTop := ctrl.Bottom() + ctx.Theme.SpaceS
			errRect := s.base.ErrorRect(ctx.Theme)
			if s.base.Invalid && s.base.ErrorText != "" {
				listTop = errRect.Bottom() + ctx.Theme.SpaceS
			}
			rowH := ctx.Theme.ControlH
			n := len(s.options)
			max := s.MaxVisible
			if max <= 0 {
				max = 5
			}
			if n > max {
				n = max
			}
			listRect := Rect{X: ctrl.X, Y: listTop, W: ctrl.W, H: n * rowH}
			if listRect.Contains(e.X, e.Y) {
				row := (e.Y - listRect.Y) / rowH
				idx := s.scroll + row
				if idx >= 0 && idx < len(s.options) {
					s.index = idx
				}
				s.open = false
			} else {
				s.open = false
			}
		}
	}
}

func (s *Select) Update(ctx *Context) {
	s.theme = ctx.Theme
	if s.base.Rect.H == 0 {
		s.SetRectByWidth(s.base.Rect.X, s.base.Rect.Y, s.base.Rect.W)
	}
	if !s.base.Enabled {
		return
	}

	// Wheel scroll on open dropdown (desktop)
	if s.open {
		_, wy := ebiten.Wheel()
		if wy != 0 {
			step := int(math.Copysign(1, wy))
			s.scroll -= step
			maxScroll := len(s.options) - s.MaxVisible
			if maxScroll < 0 {
				maxScroll = 0
			}
			if s.scroll < 0 {
				s.scroll = 0
			}
			if s.scroll > maxScroll {
				s.scroll = maxScroll
			}
		}
	}
}

func (s *Select) Draw(ctx *Context, dst *ebiten.Image) {
	s.theme = ctx.Theme
	if s.base.Rect.H == 0 {
		s.SetRectByWidth(s.base.Rect.X, s.base.Rect.Y, s.base.Rect.W)
	}

	ctrl := s.base.ControlRect(ctx.Theme)

	bg := ctx.Theme.Surface
	if !s.base.Enabled {
		bg = ctx.Theme.SurfacePressed
	} else if s.base.pressed {
		bg = ctx.Theme.SurfacePressed
	} else if s.base.hovered {
		bg = ctx.Theme.SurfaceHover
	}
	drawRoundedRect(dst, ctrl, ctx.Theme.Radius, bg)

	borderCol := ctx.Theme.Border
	if s.base.Invalid {
		borderCol = ctx.Theme.ErrorBorder
	}
	drawRoundedBorder(dst, ctrl, ctx.Theme.Radius, ctx.Theme.BorderW, borderCol)

	if s.base.focused && s.base.Enabled {
		drawFocusRing(dst, ctrl, ctx.Theme.Radius, ctx.Theme.FocusRingGap, ctx.Theme.FocusRingW, ctx.Theme.Focus)
	}

	// Value text
	met, _ := MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
	baselineY := ctrl.Y + (ctrl.H-met.Height)/2 + met.Ascent

	val := s.Value()
	if val == "" {
		val = "—"
	}
	ctx.Text.SetColor(ctx.Theme.Text)
	ctx.Text.SetAlign(0)
	ctx.Text.Draw(dst, val, ctrl.X+ctx.Theme.PadX, baselineY)

	// Chevron (simple)
	chev := "." // ▾
	cw := MeasureStringPx(ctx.Theme.Font, ctx.Theme.FontPx, chev)
	ctx.Text.Draw(dst, chev, ctrl.Right()-ctx.Theme.PadX-cw, baselineY)

	// Dropdown list
	if s.open {
		n := len(s.options)
		max := s.MaxVisible
		if max <= 0 {
			max = 5
		}
		if n > max {
			n = max
		}
		listY := ctrl.Bottom() + ctx.Theme.SpaceS
		if s.base.Invalid && s.base.ErrorText != "" {
			errRect := s.base.ErrorRect(ctx.Theme)
			listY = errRect.Bottom() + ctx.Theme.SpaceS
		}
		list := Rect{X: ctrl.X, Y: listY, W: ctrl.W, H: n * ctx.Theme.ControlH}

		drawRoundedRect(dst, list, ctx.Theme.Radius, ctx.Theme.Surface)
		drawRoundedBorder(dst, list, ctx.Theme.Radius, ctx.Theme.BorderW, ctx.Theme.Border)

		for i := 0; i < n; i++ {
			idx := s.scroll + i
			if idx >= len(s.options) {
				break
			}
			row := Rect{X: list.X, Y: list.Y + i*ctx.Theme.ControlH, W: list.W, H: ctx.Theme.ControlH}

			if idx == s.index {
				drawRoundedRect(dst, row, 0, ctx.Theme.SurfaceHover)
			}
			bY := row.Y + (row.H-met.Height)/2 + met.Ascent
			ctx.Text.SetColor(ctx.Theme.Text)
			ctx.Text.Draw(dst, s.options[idx], row.X+ctx.Theme.PadX, bY)
		}
	}

	// Validation
	err := s.base.ErrorRect(ctx.Theme)
	if s.base.Invalid {
		drawErrorText(ctx, dst, err, s.base.ErrorText)
	}
}

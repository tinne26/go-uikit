package widget

import (
	"image"

	"github.com/erparts/go-uikit/common"
	"github.com/erparts/go-uikit/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// Container is an empty widget that lets you render custom content inside a themed box.
// It still participates in focus/invalid layout like any other widget.
type Container struct {
	base ui.Base

	OnUpdate func(ctx *ui.Context, content image.Rectangle)
	OnDraw   func(ctx *ui.Context, dst *ebiten.Image, content image.Rectangle)
}

func NewContainer(theme *ui.Theme) *Container {
	cfg := ui.NewWidgetBaseConfig(theme)

	return &Container{
		base: ui.NewBase(cfg),
	}
}

func (c *Container) Base() *ui.Base  { return &c.base }
func (c *Container) Focusable() bool { return false }

func (c *Container) SetFrame(x, y, w int) {
	c.base.SetFrame(x, y, w)
}
func (c *Container) Measure() image.Rectangle { return c.base.Rect }

func (c *Container) Update(ctx *ui.Context) {
	if c.base.Rect.Dy() == 0 {
		c.base.SetFrame(c.base.Rect.Min.X, c.base.Rect.Min.Y, c.base.Rect.Dx())
	}

	if c.OnUpdate != nil {
		c.OnUpdate(ctx, common.Inset(c.base.ControlRect(ctx.Theme), ctx.Theme.PadX, ctx.Theme.PadY))
	}
}

func (c *Container) Draw(ctx *ui.Context, dst *ebiten.Image) {
	r := c.base.Draw(ctx, dst)

	content := common.Inset(r, ctx.Theme.PadX, ctx.Theme.PadY)
	if c.OnDraw != nil {
		c.OnDraw(ctx, dst, content)
	}
}

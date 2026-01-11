package demo

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"

	"github.com/erparts/go-uikit/ui"
)

type Game struct {
	ime ui.IMEBridge

	winW, winH int

	scale    ui.Scale
	renderer *etxt.Renderer
	theme    *ui.Theme
	ctx      *ui.Context

	title     *ui.Label
	txtA      *ui.TextInput
	txtB      *ui.TextInput
	txtDis    *ui.TextInput
	ta        *ui.TextArea
	sel       *ui.Select
	box       *ui.Container
	chkA      *ui.Checkbox
	chkDis    *ui.Checkbox
	btnA      *ui.Button
	btnDis    *ui.Button
	footer    *ui.Label
	focusInfo *ui.Label
}

func mustFont() *sfnt.Font {
	f, err := sfnt.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	return f
}

func New() *Game { return &Game{} }

// SetIMEBridge can be called from mobile bindings to enable keyboard show/hide.
func (g *Game) SetIMEBridge(b ui.IMEBridge) {
	g.ime = b
	if g.ctx != nil {
		g.ctx.SetIMEBridge(b)
	}
}

func (g *Game) initOnce() {
	if g.renderer != nil {
		return
	}

	g.renderer = etxt.NewRenderer()
	g.renderer.Utils().SetCache8MiB()
	g.renderer.SetAlign(etxt.Left)

	f := mustFont()
	g.renderer.SetFont(f)

	// Base theme in logical pixels. Actual rendering scale is handled by renderer.SetScale.
	g.theme = ui.NewTheme(f, 20)

	g.ctx = ui.NewContext(g.theme, g.renderer, g.ime)

	g.title = ui.NewLabel("UI Kit Demo — consistent proportions (Theme-driven)")
	g.focusInfo = ui.NewLabel("")

	g.txtA = ui.NewTextInput("Type here…")
	g.txtA.SetDefault("Hello Ebiten UI")

	g.txtB = ui.NewTextInput("Search…")

	g.txtDis = ui.NewTextInput("Disabled input")
	g.txtDis.SetDefault("Disabled value")
	g.txtDis.SetEnabled(false)

	g.ta = ui.NewTextArea("Multi-line…")
	g.ta.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7")
	g.ta.SetLines(5)

	g.sel = ui.NewSelect([]string{"Option A", "Option B", "Option C", "Option D", "Option E", "Option F"})

	g.box = ui.NewContainer()
	g.box.OnDraw = func(ctx *ui.Context, dst *ebiten.Image, content ui.Rect) {
		ctx.Text.SetColor(ctx.Theme.MutedText)
		ctx.Text.Draw(dst, "Custom content container", content.X, content.Y+content.H/2)
	}

	g.chkA = ui.NewCheckbox("Enable main button")
	g.chkA.SetChecked(true)

	g.chkDis = ui.NewCheckbox("Disabled checkbox")
	g.chkDis.SetChecked(true)
	g.chkDis.SetEnabled(false)

	g.btnA = ui.NewButton("Action (enabled)")
	g.btnA.OnClick = func() {
		g.footer.SetText("Button clicked!")
	}

	g.btnDis = ui.NewButton("Action (disabled)")
	g.btnDis.SetEnabled(false)

	g.footer = ui.NewLabel("")

	g.ctx.Add(g.title)
	g.ctx.Add(g.focusInfo)
	g.ctx.Add(g.txtA)
	g.ctx.Add(g.txtB)
	g.ctx.Add(g.txtDis)
	g.ctx.Add(g.ta)
	g.ctx.Add(g.sel)
	g.ctx.Add(g.box)
	g.ctx.Add(g.chkA)
	g.ctx.Add(g.chkDis)
	g.ctx.Add(g.btnA)
	g.ctx.Add(g.btnDis)
	g.ctx.Add(g.footer)
}

func (g *Game) Layout(outW, outH int) (int, int) {
	g.initOnce()

	g.winW, g.winH = outW, outH

	// On mobile, Ebiten input coordinates are already in this logical space.
	// Use renderer.SetScale to handle HiDPI.
	dev := ebiten.DeviceScaleFactor()
	if dev <= 0 {
		dev = 1
	}

	// Optional: make UI larger on visibly high-res screens where dev is 1.
	// This is a conservative heuristic.
	uiScale := 1.0
	minSide := float64(outW)
	if float64(outH) < minSide {
		minSide = float64(outH)
	}
	if dev == 1 && minSide >= 900 {
		uiScale = 2.0
	}

	g.scale = ui.Scale{Device: dev, UI: uiScale}
	g.ctx.SetScale(g.scale)
	g.renderer.SetScale(1)

	return outW, outH
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	padding := 12
	x := padding
	y := padding
	w := g.winW - padding*2

	g.title.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceS

	fw := g.ctx.Focused()
	if fw == nil {
		g.focusInfo.SetText("Focused: (none) — tap a widget or TAB")
	} else {
		g.focusInfo.SetText(fmt.Sprintf("Focused: %T", fw))
	}
	g.focusInfo.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceM

	g.txtA.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceS

	g.txtB.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceS

	g.txtDis.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceM

	g.ta.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceM

	g.sel.SetRectByWidth(x, y, w)
	y += g.sel.Base().Rect.H + g.theme.SpaceM

	g.box.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceM

	g.chkA.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceS

	g.chkDis.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceM

	g.txtB.Base().Invalid = true
	g.txtB.Base().ErrorText = "Invalid feedback"
	g.sel.Base().Invalid = true
	g.sel.Base().ErrorText = "Choose a valid option"

	g.btnA.SetEnabled(g.chkA.Checked())
	g.btnA.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceS

	g.btnDis.SetRectByWidth(x, y, w)
	y += g.theme.ControlH + g.theme.SpaceM

	g.footer.SetRectByWidth(x, y, w)

	g.ctx.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 22, 26, 255})
	g.ctx.Draw(screen)
}

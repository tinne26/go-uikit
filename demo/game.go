package demo

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"

	"github.com/erparts/go-uikit"
	"github.com/erparts/go-uikit/layout"
	"github.com/erparts/go-uikit/widget"
)

type Game struct {
	stack *layout.Stack
	grid  *layout.Grid
	ime   uikit.IMEBridge

	scale    uikit.Scale
	renderer *etxt.Renderer
	theme    *uikit.Theme
	ctx      *uikit.Context

	title        *widget.Label
	txtA         *widget.TextInput
	txtB         *widget.TextInput
	txtDis       *widget.TextInput
	ta           *widget.TextArea
	sel          *widget.Select
	box          *widget.Container
	chkA         *widget.Checkbox
	chkDis       *widget.Checkbox
	chkGrid      *widget.Checkbox
	btnA         *widget.Button
	btnDis       *widget.Button
	footer       *widget.Label
	focusInfo    *widget.Label
	exampleLabel *widget.Label

	clickCount int
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
func (g *Game) SetIMEBridge(b uikit.IMEBridge) {
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
	g.renderer.Glyph().SetMissHandler(etxt.OnMissNotdef)

	f := mustFont()
	g.renderer.SetFont(f)

	g.theme = uikit.NewTheme(f, 20)

	root := layout.NewStack(g.theme)
	g.ctx = uikit.NewContext(g.theme, root, g.renderer, g.ime)
	g.stack = layout.NewStack(g.theme)

	g.grid = layout.NewGrid(g.theme)
	g.grid.SetVisible(false)

	g.title = widget.NewLabel(g.theme, "")
	g.title.SetTextFunc(func() string {
		return fmt.Sprintf("UI Kit Demo (TPS: %0.2f - FPS: %0.2f)", ebiten.ActualTPS(), ebiten.ActualFPS())
	})

	g.focusInfo = widget.NewLabel(g.theme, "")
	g.focusInfo.SetTextFunc(func() string {
		if fw := g.ctx.Focused(); fw != nil {
			return fmt.Sprintf("Focused: %T", fw)
		}

		return "Focused: (none) — tap a widget or TAB"
	})

	g.exampleLabel = widget.NewLabel(g.theme, "Label example: static helper text")

	g.txtA = widget.NewTextInput(g.theme, "Type here…")

	g.txtB = widget.NewTextInput(g.theme, "Search…")
	g.txtB.On(uikit.EventValueChange, func(e uikit.Event) bool {
		v := e.Widget.(*widget.TextInput).Text()
		if v == "" {
			g.txtB.SetInvalid("Required")
			return false
		}

		g.txtB.ClearInvalid()
		return false
	}, false)

	g.txtDis = widget.NewTextInput(g.theme, "Disabled input")
	g.txtDis.SetText("Disabled value")
	g.txtDis.SetEnabled(false)

	g.ta = widget.NewTextArea(g.theme, "Multi-line…")
	g.ta.SetLines(5)
	g.ta.SetText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7")

	g.sel = widget.NewSelect(g.theme, nil)
	g.sel.SetOptions([]widget.SelectOption{
		{0, "Select a value..."},
		{1, "Option A"}, {2, "Option B"}, {3, "Option C"},
		{4, "Option D"}, {5, "Option E"}, {6, "Option F"},
	})

	g.sel.On(uikit.EventValueChange, func(e uikit.Event) bool {
		s, isSelected := g.sel.Selected()
		if s.Value == 0 || !isSelected {
			g.sel.SetInvalid("Requried value")
			return false
		}

		g.sel.ClearInvalid()
		return false
	}, false)

	g.box = widget.NewContainer(g.theme)
	g.box.SetHeight(140)
	g.box.OnDraw = func(ctx *uikit.Context, dst *ebiten.Image, content image.Rectangle) {
		s, _ := g.sel.Selected()
		lines := []string{
			"Custom container (user content)",
			fmt.Sprintf("Click Count: %d", g.clickCount),
			fmt.Sprintf("Select Value: %s ", s.Label),
			fmt.Sprintf("Search Text: %s", g.txtB.Text()),
			fmt.Sprintf("TextArea Chars: %d", len([]rune(g.ta.Text()))),
		}

		dst = dst.SubImage(content).(*ebiten.Image)

		met, _ := uikit.MetricsPx(ctx.Theme.Font, ctx.Theme.FontPx)
		y := (content.Min.Y) + met.Ascent
		x := content.Min.X

		ctx.Text.SetColor(ctx.Theme.MutedText)
		for _, ln := range lines {
			ctx.Text.Draw(dst, ln, x, y)
			y += met.Height
		}
	}

	g.chkA = widget.NewCheckbox(g.theme, "Enable main button")
	g.chkA.SetChecked(true)
	g.chkA.On(uikit.EventValueChange, func(e uikit.Event) bool {
		g.btnA.SetEnabled(e.Widget.(*widget.Checkbox).Checked())
		return false
	}, false)

	g.chkDis = widget.NewCheckbox(g.theme, "Disabled checkbox")
	g.chkDis.SetChecked(true)
	g.chkDis.SetEnabled(false)

	g.chkGrid = widget.NewCheckbox(g.theme, "Use grid layout")
	g.chkGrid.On(uikit.EventValueChange, func(e uikit.Event) bool {
		if e.Widget.(*widget.Checkbox).Checked() {
			g.grid.SetVisible(true)
			g.stack.SetVisible(false)
			return false
		}

		g.stack.SetVisible(true)
		g.grid.SetVisible(false)
		return false
	}, false)

	g.btnA = widget.NewButton(g.theme, "Action (enabled)")
	g.btnA.On(uikit.EventClick, func(_ uikit.Event) bool {
		g.clickCount++
		g.footer.SetText("Button clicked!")
		return false
	}, false)

	g.btnDis = widget.NewButton(g.theme, "Action (disabled)")
	g.btnDis.SetEnabled(false)

	g.footer = widget.NewLabel(g.theme, "")

	g.ctx.Add(g.title)
	g.ctx.Add(g.focusInfo)
	g.ctx.Add(g.chkGrid)

	g.ctx.Add(g.stack)
	g.ctx.Add(g.grid)
	g.ctx.Add(g.footer)

	contentWidgets := []uikit.Widget{
		g.exampleLabel,
		g.txtA,
		g.txtB,
		g.txtDis,
		g.ta,
		g.sel,
		g.box,
		g.chkA,
		g.chkDis,
		g.btnA,
		g.btnDis,
	}

	g.stack.SetChildren(contentWidgets)
	g.grid.SetChildren(contentWidgets)
}

func (g *Game) Update() error {
	g.ctx.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.ctx.Draw(screen)
}

func (g *Game) Layout(outW, outH int) (int, int) {
	g.initOnce()

	m := ebiten.Monitor()
	if m == nil {
		return 0, 0
	}

	dev := m.DeviceScaleFactor()
	if dev <= 0 {
		dev = 1
	}

	// Optional: make UI a bit larger on small screens.
	uiScale := 1.0
	minSide := float64(outW)
	if float64(outH) < minSide {
		minSide = float64(outH)
	}
	if minSide <= 520 {
		uiScale = 1.5
	} else if minSide <= 720 {
		uiScale = 1.25
	}

	g.scale = uikit.Scale{Device: dev, UI: uiScale}
	g.ctx.SetScale(g.scale)
	g.renderer.SetScale(1)
	return outW, outH
}

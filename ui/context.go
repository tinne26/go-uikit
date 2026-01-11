package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
)

// Context holds shared state for all widgets.
type Context struct {
	Theme *Theme
	Text  *etxt.Renderer
	IME   IMEBridge // optional (nil on desktop)
	Scale Scale     // kept for apps, but input + layout are in logical pixels

	widgets []Widget
	focus   int // -1 means none

	// Pointer state in *logical pixels* (Ebiten's standard coordinate space).
	ptrX, ptrY  int
	ptrDown     bool
	ptrJustDown bool
	ptrJustUp   bool
	ptrIsTouch  bool
	activeTouch ebiten.TouchID
	hasTouch    bool

	// Touch tracking (robust across Ebiten versions/platforms)
	prevTouches map[ebiten.TouchID]struct{}
}

func NewContext(theme *Theme, renderer *etxt.Renderer, ime IMEBridge) *Context {
	// Ensure renderer style is consistent with the theme.
	renderer.SetFont(theme.Font)
	renderer.SetSize(float64(theme.FontPx))
	return &Context{
		Theme:       theme,
		Text:        renderer,
		IME:         ime,
		Scale:       Scale{Device: 1, UI: 1},
		focus:       -1,
		prevTouches: map[ebiten.TouchID]struct{}{},
	}
}

// SetScale stores a scale value for the app. The default Context input space is logical pixels.
func (c *Context) SetScale(s Scale) { c.Scale = s }

// SetIMEBridge sets/updates the IME bridge at runtime.
// It also synchronizes the IME visibility with the currently focused widget.
func (c *Context) SetIMEBridge(b IMEBridge) {
	c.IME = b
	c.updateIMEForce(c.Focused())
}

func (c *Context) Add(w Widget) { c.widgets = append(c.widgets, w) }
func (c *Context) Widgets() []Widget {
	return c.widgets
}

func (c *Context) Focused() Widget {
	if c.focus < 0 || c.focus >= len(c.widgets) {
		return nil
	}
	return c.widgets[c.focus]
}

func (c *Context) dispatch(w Widget, e Event) {
	if w == nil {
		return
	}
	if h, ok := any(w).(EventHandler); ok {
		h.HandleEvent(c, e)
	}
}

func (c *Context) SetFocus(w Widget) {
	old := c.Focused()

	// Resolve new focus index (or -1).
	newIdx := -1
	if w != nil {
		for i, ww := range c.widgets {
			if ww == w {
				newIdx = i
				break
			}
		}
	}

	// Emit focus events if changed
	if old != nil && (newIdx != c.focus) {
		c.dispatch(old, Event{Type: EventFocusLost})
	}
	c.focus = newIdx
	newW := c.Focused()
	if newW != nil && newW != old {
		c.dispatch(newW, Event{Type: EventFocusGained})
	}

	// IME show/hide based on focused widget.
	c.updateIME(old, newW)
}

func (c *Context) updateIME(oldW, newW Widget) {
	if c.IME == nil {
		return
	}

	oldWants := false
	if oldW != nil {
		if wi, ok := any(oldW).(WantsIME); ok && wi.WantsIME() {
			oldWants = true
		}
	}
	newWants := false
	if newW != nil {
		if wi, ok := any(newW).(WantsIME); ok && wi.WantsIME() {
			newWants = true
		}
	}

	// Only issue calls on state transitions.
	if oldWants && !newWants {
		c.IME.Hide()
	}
	if !oldWants && newWants {
		c.IME.Show()
	}
}

func (c *Context) updateIMEForce(focused Widget) {
	if c.IME == nil {
		return
	}
	wants := false
	if focused != nil {
		if wi, ok := any(focused).(WantsIME); ok && wi.WantsIME() {
			wants = true
		}
	}
	if wants {
		c.IME.Show()
	} else {
		c.IME.Hide()
	}
}

func (c *Context) focusNext() {
	if len(c.widgets) == 0 {
		c.SetFocus(nil)
		return
	}
	start := c.focus
	for i := 0; i < len(c.widgets); i++ {
		idx := (start + 1 + i) % len(c.widgets)
		if c.widgets[idx].Base().Visible && c.widgets[idx].Base().Enabled && c.widgets[idx].Focusable() {
			c.SetFocus(c.widgets[idx])
			return
		}
	}
}

func (c *Context) focusPrev() {
	if len(c.widgets) == 0 {
		c.SetFocus(nil)
		return
	}
	start := c.focus
	for i := 0; i < len(c.widgets); i++ {
		idx := start - 1 - i
		for idx < 0 {
			idx += len(c.widgets)
		}
		if c.widgets[idx].Base().Visible && c.widgets[idx].Base().Enabled && c.widgets[idx].Focusable() {
			c.SetFocus(c.widgets[idx])
			return
		}
	}
}

func (c *Context) readPointerSnapshot() {
	c.ptrJustDown = false
	c.ptrJustUp = false
	c.ptrIsTouch = false

	// Touch tracking (prefer this on mobile; CursorPosition is always (0,0) there).
	touches := ebiten.TouchIDs()
	curr := map[ebiten.TouchID]struct{}{}
	for _, id := range touches {
		curr[id] = struct{}{}
	}

	// Determine transitions
	var justPressed []ebiten.TouchID
	var justReleased []ebiten.TouchID
	for id := range curr {
		if _, ok := c.prevTouches[id]; !ok {
			justPressed = append(justPressed, id)
		}
	}
	for id := range c.prevTouches {
		if _, ok := curr[id]; !ok {
			justReleased = append(justReleased, id)
		}
	}
	c.prevTouches = curr

	// Acquire an active touch when pressed
	if !c.hasTouch && len(justPressed) > 0 {
		c.activeTouch = justPressed[0]
		c.hasTouch = true
		c.ptrJustDown = true
	}

	if c.hasTouch {
		// Is it still down?
		if _, ok := curr[c.activeTouch]; ok {
			c.ptrDown = true
			c.ptrIsTouch = true
			c.ptrX, c.ptrY = ebiten.TouchPosition(c.activeTouch)
		} else {
			// Released this tick?
			c.ptrDown = false
			c.ptrIsTouch = true
			c.ptrJustUp = true
			c.hasTouch = false
		}
		return
	}

	// Mouse fallback (desktop)
	c.ptrX, c.ptrY = ebiten.CursorPosition()
	c.ptrDown = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	c.ptrJustDown = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	c.ptrJustUp = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
}

func (c *Context) Update() {
	c.readPointerSnapshot()

	// Keyboard focus traversal (desktop)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			c.focusPrev()
		} else {
			c.focusNext()
		}
	}

	for _, w := range c.widgets {
		b := w.Base()
		if !b.Visible {
			continue
		}

		inside := b.Rect.Contains(c.ptrX, c.ptrY)

		// Hover only makes sense for mouse
		b.hovered = inside && b.Enabled && !c.ptrIsTouch

		if c.ptrJustDown && inside && b.Enabled {
			b.pressed = true
			c.dispatch(w, Event{Type: EventPointerDown, X: c.ptrX, Y: c.ptrY})
			if w.Focusable() {
				c.SetFocus(w)
			}
		}

		b.focused = (c.Focused() == w) && b.Enabled && w.Focusable()

		w.Update(c)

		if c.ptrJustUp {
			if b.pressed {
				c.dispatch(w, Event{Type: EventPointerUp, X: c.ptrX, Y: c.ptrY})
				if inside && b.Enabled {
					c.dispatch(w, Event{Type: EventClick, X: c.ptrX, Y: c.ptrY})
				}
			}
			b.pressed = false
		}

		b.focused = (c.Focused() == w) && b.Enabled && w.Focusable()
	}
}

func (c *Context) Draw(dst *ebiten.Image) {
	for _, w := range c.widgets {
		if !w.Base().Visible {
			continue
		}
		w.Draw(c, dst)
	}
}

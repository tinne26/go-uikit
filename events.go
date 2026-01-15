package uikit

import "github.com/hajimehoshi/ebiten/v2"

// EventType represents high-level UI events dispatched by the UI Context.
//
// These events are intentionally kept generic so they can be used across widgets
// without depending on platform-specific details.
type EventType int

const (
	// EventNone is a zero/default value and should generally not be dispatched.
	EventNone EventType = iota
	// EventFocusGained is fired when a widget becomes the focused element.
	EventFocusGained
	// EventFocusLost is fired when a widget stops being the focused element.
	EventFocusLost
	// EventPointerDown is fired when a pointer (mouse/touch) is pressed down
	// over a widget. The event carries pointer coordinates in pixels.
	EventPointerDown
	// EventPointerUp is fired when a pointer (mouse/touch) is released.
	// The event carries pointer coordinates in pixels.
	EventPointerUp
	// EventClick is fired when a full click/tap gesture is detected
	// (typically PointerDown followed by PointerUp on the same widget).
	// The event carries pointer coordinates in pixels.
	EventClick
	// EventKeyDown is fired when a keyboard key is pressed while the widget
	// is focused. The event carries the key code (as int).
	EventKeyDown
	// EventKeyUp is fired when a keyboard key is released while the widget
	// is focused. The event carries the key code (as int).
	EventKeyUp
	// EventValueChange is fired when a widget's value changes due to user
	// interaction (e.g. text changed, checkbox toggled, slider moved, select
	// changed).
	EventValueChange
)

// Event is a UI event routed to a widget.
type Event struct {
	Widget  Widget
	Type    EventType
	Pointer *PointerStatus
	Key     ebiten.Key
}

// EventHandler is a function invoked when an event is dispatched.
// If it returns true, the event is considered handled and propagation stops.
type EventHandler func(Event) bool

// EventDispatcher is a small event router that stores handlers per EventType
// and dispatches incoming events to them in registration order.
type EventDispatcher struct {
	handlers map[EventType][]EventHandler
}

// NewEventDispatcher creates and returns a new EventDispatcher with no handlers
// registered.
func NewEventDispatcher() EventDispatcher {
	return EventDispatcher{
		handlers: make(map[EventType][]EventHandler),
	}
}

// On registers a new handler for the given event type.
//
// Handlers are executed in the order they are registered.
// If clear is true, any previously registered handlers for the same type
// are removed before adding this one.
func (d *EventDispatcher) On(t EventType, h EventHandler, clear bool) {
	if clear {
		d.handlers[t] = []EventHandler{}
	}
	d.handlers[t] = append(d.handlers[t], h)
}

// Dispatch sends the given event to all handlers registered for its EventType.
//
// Handlers are invoked sequentially in registration order.
// If a handler returns true, the event propagation stops immediately.
func (d *EventDispatcher) Dispatch(e Event) {
	for _, h := range d.handlers[e.Type] {
		if h(e) {
			break
		}
	}
}

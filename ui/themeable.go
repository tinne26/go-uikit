package ui

// Themeable is implemented by widgets that need Theme for sizing (SetFrame) or drawing.
// Layouts should call SetTheme before calling SetFrame.
type Themeable interface {
	SetTheme(theme *Theme)
}

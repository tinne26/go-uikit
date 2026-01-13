package ui

// Layout is a Widget that owns children.
type Layout interface {
	Widget

	Children() []Widget
	SetChildren([]Widget)
	Add(...Widget)
	Clear()
}

package common

import "image"

func Inset(r image.Rectangle, pxX, pxY int) image.Rectangle {
	minX := r.Min.X + pxX
	minY := r.Min.Y + pxY
	maxX := r.Max.X - pxX
	maxY := r.Max.Y - pxY

	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}

	return image.Rect(minX, minY, maxX, maxY)
}

func ChangeRectangleHeight(r image.Rectangle, h int) image.Rectangle {
	return image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+h)
}

package canvas

type Backend interface {
	ClearRect(x, y, w, h int)
}

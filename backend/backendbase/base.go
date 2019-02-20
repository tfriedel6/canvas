package backendbase

import "image/color"

type Style struct {
	Color       color.RGBA
	GlobalAlpha float64
	// radialGradient *RadialGradient
	// linearGradient *LinearGradient
	// image          *Image
}

type Shadow struct {
	Color   color.RGBA
	OffsetX float64
	OffsetY float64
	Blur    float64
}

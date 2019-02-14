package backendbase

import "image/color"

type Style struct {
	Color       color.RGBA
	GlobalAlpha float64
	// radialGradient *RadialGradient
	// linearGradient *LinearGradient
	// image          *Image
}

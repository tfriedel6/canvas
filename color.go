package canvas

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

type glColor struct {
	r, g, b, a float64
}

func colorGoToGL(color color.Color) glColor {
	ir, ig, ib, ia := color.RGBA()
	var c glColor
	c.r = float64(ir) / 65535
	c.g = float64(ig) / 65535
	c.b = float64(ib) / 65535
	c.a = float64(ia) / 65535
	return c
}

func colorGLToGo(c glColor) color.Color {
	if c.r < 0 {
		c.r = 0
	} else if c.r > 1 {
		c.r = 1
	}
	if c.g < 0 {
		c.g = 0
	} else if c.g > 1 {
		c.g = 1
	}
	if c.b < 0 {
		c.b = 0
	} else if c.b > 1 {
		c.b = 1
	}
	if c.a < 0 {
		c.a = 0
	} else if c.a > 1 {
		c.a = 1
	}
	return color.RGBA{
		R: uint8(c.r * 255),
		G: uint8(c.g * 255),
		B: uint8(c.b * 255),
		A: uint8(c.a * 255),
	}
}

func parseHexRune(rn rune) (int, bool) {
	switch {
	case rn >= '0' && rn <= '9':
		return int(rn - '0'), true
	case rn >= 'a' && rn <= 'f':
		return int(rn-'a') + 10, true
	case rn >= 'A' && rn <= 'F':
		return int(rn-'A') + 10, true
	}
	return 0, false
}

func parseHexRunePair(rn1, rn2 rune) (int, bool) {
	i1, ok := parseHexRune(rn1)
	if !ok {
		return 0, false
	}
	i2, ok := parseHexRune(rn2)
	if !ok {
		return 0, false
	}
	return i1*16 + i2, true
}

func parseColorComponent(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v) / 255, true
	case uint:
		return float64(v) / 255, true
	case uint8:
		return float64(v) / 255, true
	case string:
		if len(v) == 0 {
			return 0, false
		}
		if v[0] == '#' {
			str := v[1:]
			if len(str) > 2 {
				return 0, false
			}
			conv, err := strconv.ParseUint(v[1:], 16, 8)
			if err != nil {
				return 0, false
			}
			return float64(conv) / 255, true
		} else if strings.ContainsRune(v, '.') {
			conv, err := strconv.ParseFloat(v, 32)
			if err != nil {
				return 0, false
			}
			if conv < 0 {
				conv = 0
			} else if conv > 1 {
				conv = 1
			}
			return float64(conv), true
		} else {
			conv, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return 0, false
			}
			return float64(conv) / 255, true
		}
	}
	return 0, false
}

func parseColor(value ...interface{}) (c glColor, ok bool) {
	if len(value) == 1 {
		switch v := value[0].(type) {
		case color.Color:
			c = colorGoToGL(v)
			ok = true
			return
		case [3]float32:
			return glColor{r: float64(v[0]), g: float64(v[1]), b: float64(v[2]), a: 1}, true
		case [4]float32:
			return glColor{r: float64(v[0]), g: float64(v[1]), b: float64(v[2]), a: float64(v[3])}, true
		case [3]float64:
			return glColor{r: v[0], g: v[1], b: v[2], a: 1}, true
		case [4]float64:
			return glColor{r: v[0], g: v[1], b: v[2], a: v[3]}, true
		case [3]int:
			return glColor{r: float64(v[0]) / 255, g: float64(v[1]) / 255, b: float64(v[2]) / 255, a: 1}, true
		case [4]int:
			return glColor{r: float64(v[0]) / 255, g: float64(v[1]) / 255, b: float64(v[2]) / 255, a: float64(v[3]) / 255}, true
		case [3]uint:
			return glColor{r: float64(v[0]) / 255, g: float64(v[1]) / 255, b: float64(v[2]) / 255, a: 1}, true
		case [4]uint:
			return glColor{r: float64(v[0]) / 255, g: float64(v[1]) / 255, b: float64(v[2]) / 255, a: float64(v[3]) / 255}, true
		case [3]uint8:
			return glColor{r: float64(v[0]) / 255, g: float64(v[1]) / 255, b: float64(v[2]) / 255, a: 1}, true
		case [4]uint8:
			return glColor{r: float64(v[0]) / 255, g: float64(v[1]) / 255, b: float64(v[2]) / 255, a: float64(v[3]) / 255}, true
		case string:
			if len(v) == 0 {
				return
			}
			if v[0] == '#' {
				str := v[1:]
				if len(str) == 3 || len(str) == 4 {
					var ir, ig, ib int
					ia := 255
					ir, ok = parseHexRune(rune(str[0]))
					if !ok {
						return
					}
					ir = ir*16 + ir
					ig, ok = parseHexRune(rune(str[1]))
					if !ok {
						return
					}
					ig = ig*16 + ig
					ib, ok = parseHexRune(rune(str[2]))
					if !ok {
						return
					}
					ib = ib*16 + ib
					if len(str) == 4 {
						ia, ok = parseHexRune(rune(str[3]))
						if !ok {
							return
						}
						ia = ia*16 + ia
					}
					return glColor{r: float64(ir) / 255, g: float64(ig) / 255, b: float64(ib) / 255, a: float64(ia) / 255}, true
				} else if len(str) == 6 || len(str) == 8 {
					var ir, ig, ib int
					ia := 255
					ir, ok = parseHexRunePair(rune(str[0]), rune(str[1]))
					if !ok {
						return
					}
					ig, ok = parseHexRunePair(rune(str[2]), rune(str[3]))
					if !ok {
						return
					}
					ib, ok = parseHexRunePair(rune(str[4]), rune(str[5]))
					if !ok {
						return
					}
					if len(str) == 8 {
						ia, ok = parseHexRunePair(rune(str[6]), rune(str[7]))
						if !ok {
							return
						}
					}
					return glColor{r: float64(ir) / 255, g: float64(ig) / 255, b: float64(ib) / 255, a: float64(ia) / 255}, true
				} else {
					return
				}
			} else {
				v = strings.Replace(v, " ", "", -1)
				var ir, ig, ib, ia int
				n, err := fmt.Sscanf(v, "rgb(%d,%d,%d)", &ir, &ig, &ib)
				if err == nil && n == 3 {
					return glColor{r: float64(ir) / 255, g: float64(ig) / 255, b: float64(ib) / 255, a: 1}, true
				}
				n, err = fmt.Sscanf(v, "rgba(%d,%d,%d,%d)", &ir, &ig, &ib, &ia)
				if err == nil && n == 4 {
					return glColor{r: float64(ir) / 255, g: float64(ig) / 255, b: float64(ib) / 255, a: float64(ia) / 255}, true
				}
			}
		}
	} else if len(value) == 3 || len(value) == 4 {
		var pr, pg, pb, pa float64
		pr, ok = parseColorComponent(value[0])
		if !ok {
			return
		}
		pg, ok = parseColorComponent(value[1])
		if !ok {
			return
		}
		pb, ok = parseColorComponent(value[2])
		if !ok {
			return
		}
		if len(value) == 4 {
			pa, ok = parseColorComponent(value[3])
			if !ok {
				return
			}
		} else {
			pa = 1
		}
		return glColor{r: pr, g: pg, b: pb, a: pa}, true
	}

	return glColor{r: 0, g: 0, b: 0, a: 1}, false
}

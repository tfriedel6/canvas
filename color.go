package canvas

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

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

func parseColorComponent(value interface{}) (uint8, bool) {
	switch v := value.(type) {
	case float32:
		return uint8(math.Floor(float64(v) * 255)), true
	case float64:
		return uint8(math.Floor(v * 255)), true
	case int:
		return uint8(v), true
	case uint:
		return uint8(v), true
	case uint8:
		return v, true
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
			return uint8(conv), true
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
			return uint8(math.Round(conv * 255.0)), true
		} else {
			conv, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return 0, false
			}
			return uint8(conv), true
		}
	}
	return 0, false
}

func parseColor(value ...interface{}) (c color.RGBA, ok bool) {
	if len(value) == 1 {
		switch v := value[0].(type) {
		case color.Color:
			r, g, b, a := v.RGBA()
			c = color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
			ok = true
			return
		case [3]float32:
			return color.RGBA{
				R: uint8(math.Floor(float64(v[0] * 255))),
				G: uint8(math.Floor(float64(v[1] * 255))),
				B: uint8(math.Floor(float64(v[2] * 255))),
				A: 255}, true
		case [4]float32:
			return color.RGBA{
				R: uint8(math.Floor(float64(v[0] * 255))),
				G: uint8(math.Floor(float64(v[1] * 255))),
				B: uint8(math.Floor(float64(v[2] * 255))),
				A: uint8(math.Floor(float64(v[3] * 255)))}, true
		case [3]float64:
			return color.RGBA{
				R: uint8(math.Floor(v[0] * 255)),
				G: uint8(math.Floor(v[1] * 255)),
				B: uint8(math.Floor(v[2] * 255)),
				A: 255}, true
		case [4]float64:
			return color.RGBA{
				R: uint8(math.Floor(v[0] * 255)),
				G: uint8(math.Floor(v[1] * 255)),
				B: uint8(math.Floor(v[2] * 255)),
				A: uint8(math.Floor(v[3] * 255))}, true
		case [3]int:
			return color.RGBA{
				R: uint8(v[0]),
				G: uint8(v[1]),
				B: uint8(v[2]),
				A: 255}, true
		case [4]int:
			return color.RGBA{
				R: uint8(v[0]),
				G: uint8(v[1]),
				B: uint8(v[2]),
				A: uint8(v[3])}, true
		case [3]uint:
			return color.RGBA{
				R: uint8(v[0]),
				G: uint8(v[1]),
				B: uint8(v[2]),
				A: 255}, true
		case [4]uint:
			return color.RGBA{
				R: uint8(v[0]),
				G: uint8(v[1]),
				B: uint8(v[2]),
				A: uint8(v[3])}, true
		case [3]uint8:
			return color.RGBA{R: v[0], G: v[1], B: v[2], A: 255}, true
		case [4]uint8:
			return color.RGBA{R: v[0], G: v[1], B: v[2], A: v[3]}, true
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
					return color.RGBA{R: uint8(ir), G: uint8(ig), B: uint8(ib), A: uint8(ia)}, true
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
					return color.RGBA{R: uint8(ir), G: uint8(ig), B: uint8(ib), A: uint8(ia)}, true
				} else {
					return
				}
			} else {
				v = strings.Replace(v, " ", "", -1)
				var ir, ig, ib, ia int
				n, err := fmt.Sscanf(v, "rgb(%d,%d,%d)", &ir, &ig, &ib)
				if err == nil && n == 3 {
					return color.RGBA{R: uint8(ir), G: uint8(ig), B: uint8(ib), A: 255}, true
				}
				n, err = fmt.Sscanf(v, "rgba(%d,%d,%d,%d)", &ir, &ig, &ib, &ia)
				if err == nil && n == 4 {
					return color.RGBA{R: uint8(ir), G: uint8(ig), B: uint8(ib), A: uint8(ia)}, true
				}
			}
		}
	} else if len(value) == 3 || len(value) == 4 {
		var pr, pg, pb, pa uint8
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
			pa = 255
		}
		return color.RGBA{R: pr, G: pg, B: pb, A: pa}, true
	}

	return color.RGBA{A: 255}, false
}

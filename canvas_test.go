package canvas_test

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/softwarebackend"
	"github.com/tfriedel6/canvas/sdlcanvas"
)

var usesw = false

func run(t *testing.T, fn func(cv *canvas.Canvas)) {
	var img *image.RGBA
	if !usesw {
		wnd, cv, err := sdlcanvas.CreateWindow(100, 100, "test")
		if err != nil {
			t.Fatalf("Failed to create window: %v", err)
			return
		}
		defer wnd.Destroy()

		gl.Disable(gl.MULTISAMPLE)

		wnd.StartFrame()

		cv.ClearRect(0, 0, 100, 100)
		fn(cv)
		img = cv.GetImageData(0, 0, 100, 100)
	} else {
		backend := softwarebackend.New(100, 100)
		cv := canvas.New(backend)

		cv.SetFillStyle("#000")
		cv.FillRect(0, 0, 100, 100)
		fn(cv)
		img = cv.GetImageData(0, 0, 100, 100)
	}

	caller, _, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("Failed to get caller")
	}

	callerFunc := runtime.FuncForPC(caller)
	if callerFunc == nil {
		t.Fatal("Failed to get caller function")
	}

	const prefix = "canvas_test.Test"
	callerFuncName := callerFunc.Name()
	callerFuncName = callerFuncName[strings.Index(callerFuncName, prefix)+len(prefix):]

	fileName := fmt.Sprintf("testdata/%s.png", callerFuncName)

	_, err := os.Stat(fileName)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to stat file \"%s\": %v", fileName, err)
	}

	if os.IsNotExist(err) {
		err = writeImage(img, fileName)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	f, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("Failed to open file \"%s\": %v", fileName, err)
	}
	defer f.Close()

	refImg, err := png.Decode(f)
	if err != nil {
		t.Fatalf("Failed to decode file \"%s\": %v", fileName, err)
	}

	if b := img.Bounds(); b.Min.X != 0 || b.Min.Y != 0 || b.Max.X != 100 || b.Max.Y != 100 {
		t.Fatalf("Image bounds must be 0,0,100,100")
	}
	if b := refImg.Bounds(); b.Min.X != 0 || b.Min.Y != 0 || b.Max.X != 100 || b.Max.Y != 100 {
		t.Fatalf("Image bounds must be 0,0,100,100")
	}

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			r1, g1, b1, a1 := img.At(x, y).RGBA()
			r2, g2, b2, a2 := refImg.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				writeImage(img, fmt.Sprintf("testdata/%s_fail.png", callerFuncName))
				t.FailNow()
			}
		}
	}
}

func writeImage(img *image.RGBA, fileName string) error {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("Failed to create file \"%s\"", fileName)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return fmt.Errorf("Failed to encode PNG")
	}
	return nil
}

func TestFillRect(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#F00")
		cv.FillRect(10, 10, 10, 10)

		cv.SetFillStyle("#F008")
		cv.FillRect(30, 10, 10, 10)

		cv.SetFillStyle(64, 96, 128, 160)
		cv.FillRect(50, 10, 10, 10)

		cv.SetFillStyle(0.5, 0.7, 0.2, 0.8)
		cv.FillRect(70, 10, 10, 10)
	})
}

func TestFillConvexPath(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#0F0")
		cv.BeginPath()
		cv.MoveTo(20, 20)
		cv.LineTo(60, 20)
		cv.LineTo(80, 80)
		cv.LineTo(40, 80)
		cv.ClosePath()
		cv.Fill()
	})
}
func TestFillConcavePath(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#0F0")
		cv.BeginPath()
		cv.MoveTo(20, 20)
		cv.LineTo(60, 20)
		cv.LineTo(60, 60)
		cv.LineTo(50, 60)
		cv.LineTo(50, 40)
		cv.LineTo(20, 40)
		cv.ClosePath()
		cv.Fill()
	})
}

func TestFillHammer(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#0F0")
		cv.BeginPath()
		cv.Translate(50, 50)
		cv.Scale(0.7, 0.7)
		cv.MoveTo(-6, 60)
		cv.LineTo(-6, -50)
		cv.LineTo(-25, -50)
		cv.LineTo(-12, -60)
		cv.LineTo(25, -60)
		cv.LineTo(25, -50)
		cv.LineTo(6, -50)
		cv.LineTo(6, 60)
		cv.ClosePath()
		cv.Fill()
	})
}

func TestDrawPath(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#00F")
		cv.SetLineJoin(canvas.Miter)
		cv.SetLineWidth(8)
		cv.BeginPath()
		cv.MoveTo(10, 10)
		cv.LineTo(30, 10)
		cv.LineTo(30, 30)
		cv.LineTo(10, 30)
		cv.ClosePath()
		cv.Stroke()

		cv.SetLineJoin(canvas.Round)
		cv.BeginPath()
		cv.MoveTo(40, 10)
		cv.LineTo(60, 10)
		cv.LineTo(60, 30)
		cv.LineTo(40, 30)
		cv.ClosePath()
		cv.Stroke()

		cv.SetLineJoin(canvas.Bevel)
		cv.BeginPath()
		cv.MoveTo(70, 10)
		cv.LineTo(90, 10)
		cv.LineTo(90, 30)
		cv.LineTo(70, 30)
		cv.ClosePath()
		cv.Stroke()

		cv.SetLineCap(canvas.Butt)
		cv.BeginPath()
		cv.MoveTo(10, 40)
		cv.LineTo(30, 40)
		cv.LineTo(30, 60)
		cv.LineTo(10, 60)
		cv.Stroke()

		cv.SetLineCap(canvas.Round)
		cv.BeginPath()
		cv.MoveTo(40, 40)
		cv.LineTo(60, 40)
		cv.LineTo(60, 60)
		cv.LineTo(40, 60)
		cv.Stroke()

		cv.SetLineCap(canvas.Square)
		cv.BeginPath()
		cv.MoveTo(70, 40)
		cv.LineTo(90, 40)
		cv.LineTo(90, 60)
		cv.LineTo(70, 60)
		cv.Stroke()
	})
}

func TestMiterLimit(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#0F0")
		cv.SetLineJoin(canvas.Miter)
		cv.SetLineWidth(2.5)
		cv.SetMiterLimit(30)
		y, step := 20.0, 4.0
		for i := 0; i < 20; i++ {
			cv.LineTo(20, y)
			y += step
			cv.LineTo(80, y)
			y += step
			step *= 0.9
		}
		cv.Stroke()
	})
}

func TestLineDash(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#0F0")
		cv.SetLineWidth(2.5)
		cv.SetLineDash([]float64{4, 6, 8})
		cv.BeginPath()
		cv.MoveTo(20, 20)
		cv.LineTo(80, 20)
		cv.LineTo(80, 80)
		cv.LineTo(50, 80)
		cv.LineTo(50, 50)
		cv.LineTo(20, 50)
		cv.ClosePath()
		cv.MoveTo(30, 30)
		cv.LineTo(70, 30)
		cv.LineTo(70, 70)
		cv.LineTo(60, 70)
		cv.LineTo(60, 40)
		cv.LineTo(30, 40)
		cv.ClosePath()
		cv.Stroke()
		ld := cv.GetLineDash()
		if ld[0] != 4 || ld[1] != 6 || ld[2] != 8 || ld[3] != 4 || ld[4] != 6 || ld[5] != 8 {
			t.Fail()
		}
	})
}

func TestLineDashOffset(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#0F0")
		cv.SetLineWidth(2.5)
		cv.SetLineDash([]float64{4, 6, 8})
		cv.SetLineDashOffset(5)
		cv.BeginPath()
		cv.MoveTo(20, 20)
		cv.LineTo(80, 20)
		cv.LineTo(80, 80)
		cv.LineTo(50, 80)
		cv.LineTo(50, 50)
		cv.LineTo(20, 50)
		cv.ClosePath()
		cv.MoveTo(30, 30)
		cv.LineTo(70, 30)
		cv.LineTo(70, 70)
		cv.LineTo(60, 70)
		cv.LineTo(60, 40)
		cv.LineTo(30, 40)
		cv.ClosePath()
		cv.Stroke()
		ld := cv.GetLineDash()
		if ld[0] != 4 || ld[1] != 6 || ld[2] != 8 || ld[3] != 4 || ld[4] != 6 || ld[5] != 8 {
			t.Fail()
		}
	})
}

func TestCurves(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#00F")
		cv.SetLineWidth(2.5)
		cv.BeginPath()
		cv.Arc(30, 30, 15, 0, 4, false)
		cv.ClosePath()
		cv.MoveTo(30, 70)
		cv.LineTo(40, 70)
		cv.ArcTo(50, 70, 50, 55, 5)
		cv.ArcTo(50, 40, 55, 40, 5)
		cv.QuadraticCurveTo(70, 40, 80, 60)
		cv.BezierCurveTo(70, 80, 60, 80, 50, 90)
		cv.Stroke()
	})
}

func TestAlpha(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#F00")
		cv.SetLineWidth(2.5)
		cv.BeginPath()
		cv.Arc(30, 30, 15, 0, 4, false)
		cv.ClosePath()
		cv.MoveTo(30, 70)
		cv.LineTo(40, 70)
		cv.ArcTo(50, 70, 50, 55, 5)
		cv.ArcTo(50, 40, 55, 40, 5)
		cv.QuadraticCurveTo(70, 40, 80, 60)
		cv.BezierCurveTo(70, 80, 60, 80, 50, 90)
		cv.Stroke()

		cv.SetStrokeStyle("#0F08")
		cv.SetLineWidth(5)
		cv.BeginPath()
		cv.MoveTo(10, 10)
		cv.LineTo(90, 90)
		cv.LineTo(90, 10)
		cv.LineTo(10, 90)
		cv.ClosePath()
		cv.Stroke()

		cv.SetGlobalAlpha(0.5)
		cv.SetStrokeStyle("#FFF8")
		cv.SetLineWidth(8)
		cv.BeginPath()
		cv.MoveTo(50, 10)
		cv.LineTo(50, 90)
		cv.Stroke()
	})
}

func TestClosePath(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#0F0")
		cv.SetLineWidth(2.5)
		cv.BeginPath()
		cv.MoveTo(20, 20)
		cv.LineTo(40, 20)
		cv.LineTo(40, 40)
		cv.LineTo(20, 40)
		cv.ClosePath()
		cv.MoveTo(60, 20)
		cv.LineTo(80, 20)
		cv.LineTo(80, 40)
		cv.LineTo(60, 40)
		cv.ClosePath()
		cv.Stroke()

		cv.SetFillStyle("#00F")
		cv.BeginPath()
		cv.MoveTo(20, 60)
		cv.LineTo(40, 60)
		cv.LineTo(40, 80)
		cv.LineTo(20, 80)
		cv.ClosePath()
		cv.MoveTo(60, 60)
		cv.LineTo(80, 60)
		cv.LineTo(80, 80)
		cv.LineTo(60, 80)
		cv.ClosePath()
		cv.Fill()
	})
}

func TestLineDash2(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#0F0")
		cv.SetLineWidth(2.5)
		cv.BeginPath()
		cv.MoveTo(20, 20)
		cv.LineTo(40, 20)
		cv.LineTo(40, 40)
		cv.LineTo(20, 40)
		cv.ClosePath()
		cv.MoveTo(60, 20)
		cv.LineTo(80, 20)
		cv.LineTo(80, 40)
		cv.LineTo(60, 40)
		cv.ClosePath()
		cv.SetLineDash([]float64{4, 4})
		cv.MoveTo(20, 60)
		cv.LineTo(40, 60)
		cv.LineTo(40, 80)
		cv.LineTo(20, 80)
		cv.ClosePath()
		cv.MoveTo(60, 60)
		cv.LineTo(80, 60)
		cv.LineTo(80, 80)
		cv.LineTo(60, 80)
		cv.ClosePath()
		cv.Stroke()
	})
}
func TestText(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFont("testdata/Roboto-Light.ttf", 48)
		cv.SetFillStyle("#F00")
		cv.FillText("A BC", 0, 46)
		cv.SetStrokeStyle("#0F0")
		cv.SetLineWidth(1)
		cv.StrokeText("D EF", 0, 90)
	})
}

func TestConvex(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#F00")
		cv.BeginPath()
		cv.MoveTo(10, 10)
		cv.LineTo(20, 10)
		cv.LineTo(20, 20)
		cv.LineTo(10, 20)
		cv.LineTo(10, 10)
		cv.ClosePath()
		cv.Fill()

		cv.SetFillStyle("#0F0")
		cv.BeginPath()
		cv.MoveTo(30, 10)
		cv.LineTo(40, 10)
		cv.LineTo(40, 15)
		cv.LineTo(35, 15)
		cv.LineTo(35, 20)
		cv.LineTo(40, 20)
		cv.LineTo(40, 25)
		cv.LineTo(30, 25)
		cv.ClosePath()
		cv.Fill()

		cv.SetFillStyle("#00F")
		cv.BeginPath()
		cv.MoveTo(50, 10)
		cv.LineTo(50, 25)
		cv.LineTo(60, 25)
		cv.LineTo(60, 20)
		cv.LineTo(55, 20)
		cv.LineTo(55, 15)
		cv.LineTo(60, 15)
		cv.LineTo(60, 10)
		cv.ClosePath()
		cv.Fill()

		cv.SetFillStyle("#FFF")
		cv.BeginPath()
		cv.MoveTo(20, 35)
		cv.LineTo(80, 35)
		cv.ArcTo(90, 35, 90, 45, 10)
		cv.LineTo(90, 80)
		cv.ArcTo(90, 90, 80, 90, 10)
		cv.LineTo(20, 90)
		cv.ArcTo(10, 90, 10, 80, 10)
		cv.LineTo(10, 45)
		cv.ArcTo(10, 35, 20, 35, 10)
		cv.ClosePath()
		cv.Fill()
	})
}

func TestConvexSelfIntersecting(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#F00")
		cv.BeginPath()
		cv.MoveTo(33.7, 31.5)
		cv.LineTo(35.4, 55.0)
		cv.LineTo(53.0, 33.5)
		cv.LineTo(56.4, 62.4)
		cv.ClosePath()
		cv.Fill()
	})
}

func TestTransform(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		path := cv.NewPath2D()
		path.MoveTo(-10, -10)
		path.LineTo(10, -10)
		path.LineTo(0, 10)
		path.ClosePath()

		cv.Translate(40, 20)
		cv.BeginPath()
		cv.LineTo(10, 10)
		cv.LineTo(30, 10)
		cv.LineTo(20, 30)
		cv.ClosePath()
		cv.SetStrokeStyle("#F00")
		cv.Stroke()
		cv.SetStrokeStyle("#0F0")
		cv.StrokePath(path)
		cv.Translate(20, 0)
		cv.SetStrokeStyle("#00F")
		cv.StrokePath(path)
		cv.Translate(-40, 30)
		cv.BeginPath()
		cv.LineTo(10, 10)
		cv.LineTo(30, 10)
		cv.LineTo(20, 30)
		cv.ClosePath()
		cv.Translate(20, 0)
		cv.SetStrokeStyle("#FF0")
		cv.Stroke()
		cv.Translate(20, 0)
		cv.SetStrokeStyle("#F0F")
		cv.StrokePath(path)
	})
}

func TestTransform2(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetStrokeStyle("#FFF")
		cv.SetLineWidth(16)
		cv.MoveTo(20, 20)
		cv.LineTo(20, 50)
		cv.Scale(2, 1)
		cv.LineTo(45, 80)
		cv.SetLineJoin(canvas.Round)
		cv.SetLineCap(canvas.Round)
		cv.Stroke()
	})
}

func TestImage(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.DrawImage("testdata/cat.jpg", 5, 5, 40, 40)
		cv.DrawImage("testdata/cat.jpg", 100, 100, 100, 100, 55, 55, 40, 40)
		cv.Save()
		cv.Translate(75, 25)
		cv.Rotate(math.Pi / 2)
		cv.Translate(-20, -20)
		cv.DrawImage("testdata/cat.jpg", 0, 0, 40, 40)
		cv.Restore()
		cv.SetTransform(1, 0, 0.2, 1, 0, 0)
		cv.DrawImage("testdata/cat.jpg", -8, 55, 40, 40)
	})
}

func TestGradient(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.Translate(50, 50)
		cv.Scale(0.8, 0.7)
		cv.Rotate(math.Pi * 0.1)
		cv.Translate(-50, -50)

		lg := cv.CreateLinearGradient(10, 10, 40, 20)
		lg.AddColorStop(0, 0.5, 0, 0)
		lg.AddColorStop(0.5, "#008000")
		lg.AddColorStop(1, 0, 0, 128)
		cv.SetFillStyle(lg)
		cv.FillRect(0, 0, 50, 100)

		rg := cv.CreateRadialGradient(75, 15, 10, 75, 75, 20)
		rg.AddColorStop(0, 1.0, 0, 0, 0.5)
		rg.AddColorStop(0.5, "#00FF0080")
		rg.AddColorStop(1, 0, 0, 255, 128)
		cv.SetFillStyle(rg)
		cv.FillRect(50, 0, 50, 100)
	})
}

func TestImagePattern(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.Translate(50, 50)
		cv.Scale(0.95, 1.05)
		cv.Rotate(-math.Pi * 0.1)

		cv.SetFillStyle("testdata/cat.jpg")
		cv.FillRect(-40, -40, 80, 80)
	})
}

func TestImagePattern2(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		ptrn := cv.CreatePattern("testdata/cat.jpg", canvas.NoRepeat)
		ptrn.SetTransform([6]float64{0, 0.1, 0.1, 0, 0, 0})

		cv.Translate(50, 50)
		cv.Scale(0.95, 1.05)
		cv.Rotate(-math.Pi * 0.1)

		cv.SetFillStyle(ptrn)
		cv.FillRect(-40, -40, 80, 80)
	})
}

func TestShadow(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		cv.SetFillStyle("#800")
		cv.SetShadowColor("#00F")
		cv.SetShadowOffset(6, 6)
		cv.FillRect(10, 10, 60, 25)
		cv.SetShadowBlur(6)
		cv.FillRect(10, 55, 60, 25)
		cv.SetFillStyle("#0F0")
		cv.SetShadowColor("#F0F")
		cv.SetGlobalAlpha(0.5)
		cv.FillRect(50, 15, 30, 60)
	})
}

func TestReadme(t *testing.T) {
	run(t, func(cv *canvas.Canvas) {
		w, h := 100.0, 100.0
		cv.SetFillStyle("#000")
		cv.FillRect(0, 0, w, h)

		for r := 0.0; r < math.Pi*2; r += math.Pi * 0.1 {
			cv.SetFillStyle(int(r*10), int(r*20), int(r*40))
			cv.BeginPath()
			cv.MoveTo(w*0.5, h*0.5)
			cv.Arc(w*0.5, h*0.5, math.Min(w, h)*0.4, r, r+0.1*math.Pi, false)
			cv.ClosePath()
			cv.Fill()
		}

		cv.SetStrokeStyle("#FFF")
		cv.SetLineWidth(10)
		cv.BeginPath()
		cv.Arc(w*0.5, h*0.5, math.Min(w, h)*0.4, 0, math.Pi*2, false)
		cv.Stroke()
	})
}

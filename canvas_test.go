package canvas_test

import (
	"fmt"
	"image/png"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas"
	_ "github.com/tfriedel6/canvas/glimpl/xmobile"
	"github.com/tfriedel6/canvas/sdlcanvas"
)

func run(t *testing.T, fn func(cv *canvas.Canvas)) {
	wnd, cv, err := sdlcanvas.CreateWindow(100, 100, "test")
	if err != nil {
		t.Fatalf("Failed to crete window: %v", err)
		return
	}
	defer wnd.Destroy()

	gl.Disable(gl.MULTISAMPLE)

	wnd.StartFrame()
	cv.ClearRect(0, 0, 100, 100)
	fn(cv)
	img := cv.GetImageData(0, 0, 100, 100)

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

	_, err = os.Stat(fileName)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to stat file \"%s\": %v", fileName, err)
	}

	if os.IsNotExist(err) {
		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
		if err != nil {
			t.Fatalf("Failed to create file \"%s\"", fileName)
		}
		defer f.Close()

		err = png.Encode(f, img)
		if err != nil {
			t.Fatalf("Failed to encode PNG")
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
				t.FailNow()
			}
		}
	}
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

		cv.SetLineEnd(canvas.Butt)
		cv.BeginPath()
		cv.MoveTo(10, 40)
		cv.LineTo(30, 40)
		cv.LineTo(30, 60)
		cv.LineTo(10, 60)
		cv.Stroke()

		cv.SetLineEnd(canvas.Round)
		cv.BeginPath()
		cv.MoveTo(40, 40)
		cv.LineTo(60, 40)
		cv.LineTo(60, 60)
		cv.LineTo(40, 60)
		cv.Stroke()

		cv.SetLineEnd(canvas.Square)
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

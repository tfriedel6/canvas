package canvas_test

import (
	"fmt"
	"image/png"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/sdlcanvas"
)

func run(t *testing.T, fn func(cv *canvas.Canvas)) {
	wnd, cv, err := sdlcanvas.CreateWindow(100, 100, "test")
	if err != nil {
		t.Fatalf("Failed to crete window: %v", err)
		return
	}
	defer wnd.Destroy()

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

	fileName := fmt.Sprintf("testimages/%s.png", callerFuncName)

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

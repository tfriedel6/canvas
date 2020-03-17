package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	{ // make sure we are in the right directory
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		d1 := filepath.Base(dir)
		d2 := filepath.Base(filepath.Dir(dir))
		if d2 != "backend" || d1 != "xmobilebackend" {
			log.Fatalln("This must be run in the backend/xmobilebackend directory")
		}
	}

	{ // delete existing files
		fis, err := ioutil.ReadDir(".")
		if err != nil {
			log.Fatalf("Failed to read current dir: %v", err)
		}

		for _, fi := range fis {
			if fi.IsDir() || fi.Name() == "shader.go" {
				continue
			}
			err = os.Remove(fi.Name())
			if err != nil {
				log.Fatalf("Failed to delete file %s: %v", fi.Name(), err)
			}
		}
	}

	{ // copy gogl files
		fis, err := ioutil.ReadDir("../goglbackend")
		if err != nil {
			log.Fatalf("Failed to read dir ../goglbackend: %v", err)
		}

		for _, fi := range fis {
			if !strings.HasSuffix(fi.Name(), ".go") || fi.Name() == "shader.go" {
				continue
			}
			path := filepath.Join("../goglbackend", fi.Name())
			data, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatalf("Failed to read file %s: %v", path, err)
			}

			filename, rewritten := rewrite(fi.Name(), string(data))

			err = ioutil.WriteFile(filename, ([]byte)(rewritten), 0777)
			if err != nil {
				log.Fatalf("Failed to write file %s: %v", fi.Name(), err)
			}
		}
	}

	err := exec.Command("go", "fmt").Run()
	if err != nil {
		log.Fatalf("Failed to run go fmt: %v", err)
	}
}

func rewrite(filename, src string) (string, string) {
	src = strings.Replace(src, `package goglbackend`, `package xmobilebackend`, 1)
	src = strings.Replace(src, `"github.com/tfriedel6/canvas/backend/goglbackend/gl"`, `"golang.org/x/mobile/gl"`, 1)
	src = strings.Replace(src, "\tgl.", "\tb.glctx.", -1)
	src = strings.Replace(src, "GoGLBackend", "XMobileBackend", -1)
	src = strings.Replace(src, "uint32(gl.TRIANGLES)", "gl.Enum(gl.TRIANGLES)", -1)

	src = strings.Replace(src, `func (g *gradient) Delete() {`,
		`func (g *gradient) Delete() {
	b := g.b`, -1)
	src = strings.Replace(src, `func (g *gradient) load(stops backendbase.Gradient) {`,
		`func (g *gradient) load(stops backendbase.Gradient) {
	b := g.b`, -1)
	src = strings.Replace(src, `func (img *Image) Delete() {`,
		`func (img *Image) Delete() {
	b := img.b`, -1)
	src = strings.Replace(src, `func (img *Image) Replace(src image.Image) error {`,
		`func (img *Image) Replace(src image.Image) error {
	b := img.b`, -1)

	src = strings.Replace(src, `imageBufTex == 0`, `imageBufTex.Value == 0`, -1)

	src = strings.Replace(src,
		`loadImage(src image.Image, tex uint32)`,
		`loadImage(b *XMobileBackend, src image.Image, tex gl.Texture)`, -1)
	src = strings.Replace(src,
		`loadImageRGBA(src *image.RGBA, tex uint32)`,
		`loadImageRGBA(b *XMobileBackend, src *image.RGBA, tex gl.Texture)`, -1)
	src = strings.Replace(src,
		`loadImageGray(src *image.Gray, tex uint32)`,
		`loadImageGray(b *XMobileBackend, src *image.Gray, tex gl.Texture)`, -1)
	src = strings.Replace(src,
		`loadImageConverted(src image.Image, tex uint32)`,
		`loadImageConverted(b *XMobileBackend, src image.Image, tex gl.Texture)`, -1)

	src = strings.Replace(src,
		`func loadShader(vs, fs string, sp *shaderProgram) error {`,
		`func loadShader(b *XMobileBackend, vs, fs string, sp *shaderProgram) error {
	sp.b = b`, -1)

	src = strings.Replace(src, `func glError() error {
	glErr := gl.GetError()
`, `func glError(b *XMobileBackend) error {
	glErr := b.glctx.GetError()
`, -1)
	src = rewriteCalls(src, "glError", func(params []string) string {
		return "glError(b)"
	})

	src = regexp.MustCompile(`[ \t]+tex[ ]+uint32`).ReplaceAllString(src, "\ttex gl.Texture")

	src = rewriteCalls(src, "b.glctx.BufferData", func(params []string) string {
		return "b.glctx.BufferData(" + params[0] + ", byteSlice(" + params[2] + ", " + params[1] + "), " + params[3] + ")"
	})
	src = rewriteCalls(src, "b.glctx.VertexAttribPointer", func(params []string) string {
		params[5] = strings.Replace(params[5], "gl.PtrOffset(", "", 1)
		params[5] = strings.Replace(params[5], ")", "", 1)
		params[5] = strings.Replace(params[5], "nil", "0", 1)
		return "b.glctx.VertexAttribPointer(" + strings.Join(params, ",") + ")"
	})
	src = rewriteCalls(src, "b.glctx.DrawArrays", func(params []string) string {
		if strings.HasPrefix(params[2], "int32(") {
			params[2] = params[2][6 : len(params[2])-1]
		}
		return "b.glctx.DrawArrays(" + strings.Join(params, ",") + ")"
	})
	src = rewriteCalls(src, "b.glctx.Uniform1fv", func(params []string) string {
		params[2] = params[2][1 : len(params[2])-3]
		return "b.glctx.Uniform1fv(" + params[0] + ", " + params[2] + ")"
	})
	src = rewriteCalls(src, "b.glctx.Uniform1i", func(params []string) string {
		if strings.HasPrefix(params[1], "int32(") {
			params[1] = params[1][6 : len(params[1])-1]
		}
		return "b.glctx.Uniform1i(" + strings.Join(params, ",") + ")"
	})
	src = rewriteCalls(src, "b.glctx.UniformMatrix3fv", func(params []string) string {
		return "b.glctx.UniformMatrix3fv(" + params[0] + ", " + params[3][1:len(params[3])-3] + "[:])"
	})
	src = rewriteCalls(src, "b.glctx.TexImage2D", func(params []string) string {
		params = append(params[:5], params[6:]...)
		for i, param := range params {
			if strings.HasPrefix(param, "int32(") {
				params[i] = param[6 : len(param)-1]
			} else if strings.HasPrefix(param, "gl.Ptr(") {
				params[i] = param[8:len(param)-2] + ":]"
			}
		}
		return "b.glctx.TexImage2D(" + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "b.glctx.TexSubImage2D", func(params []string) string {
		for i, param := range params {
			if strings.HasPrefix(param, "int32(") {
				params[i] = param[6 : len(param)-1]
			} else if strings.HasPrefix(param, "gl.Ptr(") {
				params[i] = param[8:len(param)-2] + ":]"
			}
		}
		return "b.glctx.TexSubImage2D(" + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "b.glctx.GetIntegerv", func(params []string) string {
		return "b.glctx.GetIntegerv(" + params[1][1:len(params[1])-2] + ":], " + params[0] + ")"
	})
	src = rewriteCalls(src, "b.glctx.GenTextures", func(params []string) string {
		return params[1][1:] + " = b.glctx.CreateTexture()"
	})
	src = rewriteCalls(src, "b.glctx.DeleteTextures", func(params []string) string {
		return "b.glctx.DeleteTexture(" + params[1][1:] + ")"
	})
	src = rewriteCalls(src, "b.glctx.GenBuffers", func(params []string) string {
		return params[1][1:] + " = b.glctx.CreateBuffer()"
	})
	src = rewriteCalls(src, "b.glctx.GenFramebuffers", func(params []string) string {
		return params[1][1:] + " = b.glctx.CreateFramebuffer()"
	})
	src = rewriteCalls(src, "b.glctx.DeleteFramebuffers", func(params []string) string {
		return "b.glctx.DeleteFramebuffer(" + params[1][1:] + ")"
	})
	src = rewriteCalls(src, "b.glctx.GenRenderbuffers", func(params []string) string {
		return params[1][1:] + " = b.glctx.CreateRenderbuffer()"
	})
	src = rewriteCalls(src, "b.glctx.DeleteRenderbuffers", func(params []string) string {
		return "b.glctx.DeleteRenderbuffer(" + params[1][1:] + ")"
	})
	src = rewriteCalls(src, "b.glctx.RenderbufferStorage", func(params []string) string {
		for i, param := range params {
			if strings.HasPrefix(param, "int32(") {
				params[i] = param[6 : len(param)-1]
			}
		}
		return "b.glctx.RenderbufferStorage(" + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "gl.CheckFramebufferStatus", func(params []string) string {
		return "b.glctx.CheckFramebufferStatus(" + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "b.glctx.BindFramebuffer", func(params []string) string {
		if params[1] == "0" {
			params[1] = "gl.Framebuffer{Value: 0}"
		}
		return "b.glctx.BindFramebuffer(" + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "b.glctx.ReadPixels", func(params []string) string {
		for i, param := range params {
			if strings.HasPrefix(param, "int32(") {
				params[i] = param[6 : len(param)-1]
			} else if strings.HasPrefix(param, "gl.Ptr(") {
				params[i] = param[8:len(param)-2] + ":]"
			} else if len(param) >= 5 && param[:3] == "vp[" {
				params[i] = fmt.Sprintf("int(%s)", param)
			}
		}
		return "b.glctx.ReadPixels(" + params[6] + ", " + strings.Join(params[:len(params)-1], ", ") + ")"
	})
	src = rewriteCalls(src, "b.glctx.Viewport", func(params []string) string {
		for i, param := range params {
			if strings.HasPrefix(param, "int32(") {
				params[i] = param[6 : len(param)-1]
			}
		}
		return "b.glctx.Viewport(" + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "loadImage", func(params []string) string {
		return "loadImage(b, " + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "loadImageRGBA", func(params []string) string {
		return "loadImageRGBA(b, " + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "loadImageGray", func(params []string) string {
		return "loadImageGray(b, " + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "loadImageConverted", func(params []string) string {
		return "loadImageConverted(b, " + strings.Join(params, ", ") + ")"
	})
	src = rewriteCalls(src, "loadShader", func(params []string) string {
		return "loadShader(b, " + strings.Join(params, ", ") + ")"
	})
	src = strings.ReplaceAll(src, "if tex == 0 {", "if tex.Value == 0 {")

	if filename == "gogl.go" {
		filename = "xmobile.go"
		src = rewriteMain(src)
	} else if filename == "shaders.go" {
		src = rewriteShaders(src)
	}

	return filename, src
}

func rewriteMain(src string) string {
	src = strings.Replace(src, "type GLContext struct {\n",
		"type GLContext struct {\n\tglctx gl.Context\n\n", 1)
	src = strings.Replace(src, "ctx := &GLContext{\n",
		"ctx := &GLContext{\n\t\tglctx: glctx,\n\n", 1)
	src = strings.Replace(src, "\tb.glctx.GetError() // clear error state\n",
		"\tb := &XMobileBackend{GLContext: ctx}\n\n\tb.glctx.GetError() // clear error state\n\n", 1)
	src = strings.Replace(src, "type XMobileBackend struct {\n",
		"type XMobileBackend struct {\n", 1)
	src = strings.Replace(src, "func NewGLContext() (*GLContext, error) {",
		"func NewGLContext(glctx gl.Context) (*GLContext, error) {", 1)
	src = strings.Replace(src, "TextureID uint32", "TextureID gl.Texture", 1)

	src = strings.Replace(src,
		`	err := gl.Init()
	if err != nil {
		return nil, err
	}

`, `	var err error

`, 1)

	src = strings.Replace(src,
		`// New returns a new canvas backend. x, y, w, h define the target
// rectangle in the window. ctx is a GLContext created with
// NewGLContext, but can be nil for a default one. It makes sense
// to pass one in when using for example an onscreen and an
// offscreen backend using the same GL context.
`, `// New returns a new canvas backend. x, y, w, h define the target
// rectangle in the window. ctx is a GLContext created with
// NewGLContext
`, 1)
	src = strings.Replace(src,
		`// NewOffscreen returns a new offscreen canvas backend. w, h define
// the size of the offscreen texture. ctx is a GLContext created
// with NewGLContext, but can be nil for a default one. It makes
// sense to pass one in when using for example an onscreen and an
// offscreen backend using the same GL context.
`, `// NewOffscreen returns a new offscreen canvas backend. w, h define
// the size of the offscreen texture. ctx is a GLContext created
// with NewGLContext
`, 1)
	src = strings.Replace(src,
		`	if ctx == nil {
		var err error
		ctx, err = NewGLContext()
		if err != nil {
			return nil, err
		}
	}

`, "", 1)

	src = strings.Replace(src,
		`	buf       uint32
	shadowBuf uint32
	alphaTex  uint32
`,
		`	buf       gl.Buffer
	shadowBuf gl.Buffer
	alphaTex  gl.Texture
`, 1)

	src = strings.Replace(src, `imageBufTex uint32`, `imageBufTex gl.Texture`, 1)

	src = strings.Replace(src,
		`type offscreenBuffer struct {
	tex gl.Texture
	w                int
	h                int
	renderStencilBuf uint32
	frameBuf         uint32
	alpha            bool
}
`,
		`type offscreenBuffer struct {
	tex              gl.Texture
	w                int
	h                int
	renderStencilBuf gl.Renderbuffer
	frameBuf         gl.Framebuffer
	alpha            bool
}
`, 1)

	src = src + `
func byteSlice(ptr unsafe.Pointer, size int) []byte {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = size
	sh.Len = size
	sh.Data = uintptr(ptr)
	return buf
}
`
	src = strings.Replace(src, "import (\n",
		`import (
	"unsafe"
	"reflect"
	`, 1)
	src = strings.Replace(src, "vertexLoc uint32", "vertexLoc gl.Attrib", 1)
	src = strings.Replace(src, "alphaTexSlot int32", "alphaTexSlot int", 1)
	src = strings.Replace(src, "vertexLoc, alphaTexCoordLoc uint32", "vertexLoc, alphaTexCoordLoc gl.Attrib", 1)

	return src
}

func rewriteShaders(src string) string {
	src = strings.Replace(src,
		`package xmobilebackend
`,
		`package xmobilebackend

import (
	"golang.org/x/mobile/gl"
)
`, 1)

	src = strings.Replace(src, "uint32", "gl.Attrib", -1)
	src = strings.Replace(src, "int32", "gl.Uniform", -1)
	src = strings.Replace(src, "shdFuncSolid gl.Uniform", "shdFuncSolid int", -1)

	return src
}

func rewriteCalls(src, funcName string, fn func([]string) string) string {
	rewritten := ""
	pos := 0
	for {
		idx := strings.Index(src[pos:], funcName)
		if idx == -1 {
			rewritten += src[pos:]
			break
		}
		idx += pos
		rewritten += src[pos:idx]
		parenStart := idx + len(funcName)
		if idx > 5 && src[idx-5:idx] == "func " {
			rewritten += src[idx:parenStart]
			pos = parenStart
			continue
		}
		if src[parenStart] != '(' {
			rewritten += src[idx:parenStart]
			pos = parenStart
			continue
		}

		params := make([]string, 0, 10)

		parenDepth := 0
		paramStart := 0
		paramsStr := src[parenStart+1:]
	paramloop:
		for i, rn := range paramsStr {
			switch rn {
			case '(':
				parenDepth++
			case ')':
				parenDepth--
				if parenDepth == -1 {
					params = append(params, strings.TrimSpace(paramsStr[paramStart:i]))
					pos = parenStart + i + 2
					break paramloop
				}
			case ',':
				params = append(params, strings.TrimSpace(paramsStr[paramStart:i]))
				paramStart = i + 1
			}
		}

		rewritten += fn(params)
	}

	return rewritten
}

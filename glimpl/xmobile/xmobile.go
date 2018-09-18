package glimplxmobile

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/tfriedel6/canvas"
	"golang.org/x/mobile/gl"
)

type GLImpl struct {
	gl       gl.Context
	programs map[uint32]gl.Program
}

var _ canvas.GL = GLImpl{}

func New(ctx gl.Context) *GLImpl {
	return &GLImpl{
		gl:       ctx,
		programs: make(map[uint32]gl.Program),
	}
}

func (gli GLImpl) Ptr(data interface{}) unsafe.Pointer {
	if data == nil {
		return unsafe.Pointer(nil)
	}
	var addr unsafe.Pointer
	v := reflect.ValueOf(data)
	switch v.Type().Kind() {
	case reflect.Ptr:
		e := v.Elem()
		switch e.Kind() {
		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			addr = unsafe.Pointer(e.UnsafeAddr())
		default:
			panic(fmt.Errorf("unsupported pointer to type %s; must be a slice or pointer to a singular scalar value or the first element of an array or slice", e.Kind()))
		}
	case reflect.Uintptr:
		addr = unsafe.Pointer(v.Pointer())
	case reflect.Slice:
		addr = unsafe.Pointer(v.Index(0).UnsafeAddr())
	default:
		panic(fmt.Errorf("unsupported type %s; must be a slice or pointer to a singular scalar value or the first element of an array or slice", v.Type()))
	}
	return addr
}
func (gli GLImpl) ActiveTexture(texture uint32) {
	gli.gl.ActiveTexture(gl.Enum(texture))
}
func (gli GLImpl) AttachShader(program uint32, shader uint32) {
	gli.gl.AttachShader(gli.programs[program], gl.Shader{Value: shader})
}
func (gli GLImpl) BindBuffer(target uint32, buffer uint32) {
	gli.gl.BindBuffer(gl.Enum(target), gl.Buffer{Value: buffer})
}
func (gli GLImpl) BindFramebuffer(target uint32, framebuffer uint32) {
	gli.gl.BindFramebuffer(gl.Enum(target), gl.Framebuffer{Value: framebuffer})
}
func (gli GLImpl) BindRenderbuffer(target uint32, renderbuffer uint32) {
	gli.gl.BindRenderbuffer(gl.Enum(target), gl.Renderbuffer{Value: renderbuffer})
}
func (gli GLImpl) BindTexture(target uint32, texture uint32) {
	gli.gl.BindTexture(gl.Enum(target), gl.Texture{Value: texture})
}
func (gli GLImpl) BlendFunc(sfactor uint32, dfactor uint32) {
	gli.gl.BlendFunc(gl.Enum(sfactor), gl.Enum(dfactor))
}
func (gli GLImpl) BufferData(target uint32, size int, data unsafe.Pointer, usage uint32) {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = size
	sh.Len = size
	sh.Data = uintptr(data)
	gli.gl.BufferData(gl.Enum(target), buf, gl.Enum(usage))
}
func (gli GLImpl) CheckFramebufferStatus(target uint32) uint32 {
	return uint32(gli.gl.CheckFramebufferStatus(gl.Enum(target)))
}
func (gli GLImpl) Clear(mask uint32) {
	gli.gl.Clear(gl.Enum(mask))
}
func (gli GLImpl) ClearColor(red float32, green float32, blue float32, alpha float32) {
	gli.gl.ClearColor(red, green, blue, alpha)
}
func (gli GLImpl) ColorMask(red bool, green bool, blue bool, alpha bool) {
	gli.gl.ColorMask(red, green, blue, alpha)
}
func (gli GLImpl) CompileShader(shader uint32) {
	gli.gl.CompileShader(gl.Shader{Value: shader})
}
func (gli GLImpl) CreateProgram() uint32 {
	program := gli.gl.CreateProgram()
	gli.programs[program.Value] = program
	return program.Value
}
func (gli GLImpl) CreateShader(xtype uint32) uint32 {
	return gli.gl.CreateShader(gl.Enum(xtype)).Value
}
func (gli GLImpl) DeleteShader(shader uint32) {
	gli.gl.DeleteShader(gl.Shader{Value: shader})
}
func (gli GLImpl) DeleteFramebuffers(n int32, framebuffers *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(framebuffers))
	for i := 0; i < int(n); i++ {
		gli.gl.DeleteFramebuffer(gl.Framebuffer{Value: buf[i]})
	}
}
func (gli GLImpl) DeleteRenderbuffers(n int32, renderbuffers *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(renderbuffers))
	for i := 0; i < int(n); i++ {
		gli.gl.DeleteRenderbuffer(gl.Renderbuffer{Value: buf[i]})
	}
}
func (gli GLImpl) DeleteTextures(n int32, textures *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(textures))
	for i := 0; i < int(n); i++ {
		gli.gl.DeleteTexture(gl.Texture{Value: buf[i]})
	}
}
func (gli GLImpl) Disable(cap uint32) {
	gli.gl.Disable(gl.Enum(cap))
}
func (gli GLImpl) DisableVertexAttribArray(index uint32) {
	gli.gl.DisableVertexAttribArray(gl.Attrib{Value: uint(index)})
}
func (gli GLImpl) DrawArrays(mode uint32, first int32, count int32) {
	gli.gl.DrawArrays(gl.Enum(mode), int(first), int(count))
}
func (gli GLImpl) Enable(cap uint32) {
	gli.gl.Enable(gl.Enum(cap))
}
func (gli GLImpl) EnableVertexAttribArray(index uint32) {
	gli.gl.EnableVertexAttribArray(gl.Attrib{Value: uint(index)})
}
func (gli GLImpl) FramebufferRenderbuffer(target uint32, attachment uint32, renderbuffertarget uint32, renderbuffer uint32) {
	gli.gl.FramebufferRenderbuffer(gl.Enum(target), gl.Enum(attachment), gl.Enum(renderbuffertarget), gl.Renderbuffer{Value: renderbuffer})
}
func (gli GLImpl) FramebufferTexture(target uint32, attachment uint32, texture uint32, level int32) {
	gli.gl.FramebufferTexture2D(gl.Enum(target), gl.Enum(attachment), gl.TEXTURE_2D, gl.Texture{Value: texture}, int(level))
}
func (gli GLImpl) GenBuffers(n int32, buffers *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(buffers))
	for i := 0; i < int(n); i++ {
		buf[i] = gli.gl.CreateBuffer().Value
	}
}
func (gli GLImpl) GenFramebuffers(n int32, framebuffers *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(framebuffers))
	for i := 0; i < int(n); i++ {
		buf[i] = gli.gl.CreateFramebuffer().Value
	}
}
func (gli GLImpl) GenRenderbuffers(n int32, renderbuffers *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(renderbuffers))
	for i := 0; i < int(n); i++ {
		buf[i] = gli.gl.CreateRenderbuffer().Value
	}
}
func (gli GLImpl) GenTextures(n int32, textures *uint32) {
	var buf []uint32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(n)
	sh.Len = int(n)
	sh.Data = uintptr(unsafe.Pointer(textures))
	for i := 0; i < int(n); i++ {
		buf[i] = gli.gl.CreateTexture().Value
	}
}
func (gli GLImpl) GenerateMipmap(target uint32) {
	gli.gl.GenerateMipmap(gl.Enum(target))
}
func (gli GLImpl) GetAttribLocation(program uint32, name string) int32 {
	return int32(gli.gl.GetAttribLocation(gli.programs[program], name).Value)
}
func (gli GLImpl) GetError() uint32 {
	return uint32(gli.gl.GetError())
}
func (gli GLImpl) GetProgramInfoLog(program uint32) string {
	return gli.gl.GetProgramInfoLog(gli.programs[program])
}
func (gli GLImpl) GetProgramiv(program uint32, pname uint32, params *int32) {
	i := gli.gl.GetProgrami(gli.programs[program], gl.Enum(pname))
	*params = int32(i)
}
func (gli GLImpl) GetShaderInfoLog(program uint32) string {
	return gli.gl.GetShaderInfoLog(gl.Shader{Value: program})
}
func (gli GLImpl) GetShaderiv(shader uint32, pname uint32, params *int32) {
	i := gli.gl.GetShaderi(gl.Shader{Value: shader}, gl.Enum(pname))
	*params = int32(i)
}
func (gli GLImpl) GetUniformLocation(program uint32, name string) int32 {
	return gli.gl.GetUniformLocation(gli.programs[program], name).Value
}
func (gli GLImpl) LinkProgram(program uint32) {
	gli.gl.LinkProgram(gli.programs[program])
}
func (gli GLImpl) ReadPixels(x int32, y int32, width int32, height int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(width * height * 4)
	sh.Len = int(width * height * 4)
	sh.Data = uintptr(pixels)
	gli.gl.ReadPixels(buf, int(x), int(y), int(width), int(height), gl.Enum(format), gl.Enum(xtype))
}
func (gli GLImpl) RenderbufferStorage(target uint32, internalformat uint32, width int32, height int32) {
	gli.gl.RenderbufferStorage(gl.Enum(target), gl.Enum(internalformat), int(width), int(height))
}
func (gli GLImpl) Scissor(x int32, y int32, width int32, height int32) {
	gli.gl.Scissor(x, y, width, height)
}
func (gli GLImpl) ShaderSource(shader uint32, source string) {
	gli.gl.ShaderSource(gl.Shader{Value: shader}, source)
}
func (gli GLImpl) StencilFunc(xfunc uint32, ref int32, mask uint32) {
	gli.gl.StencilFunc(gl.Enum(xfunc), int(ref), mask)
}
func (gli GLImpl) StencilMask(mask uint32) {
	gli.gl.StencilMask(mask)
}
func (gli GLImpl) StencilOp(fail uint32, zfail uint32, zpass uint32) {
	gli.gl.StencilOp(gl.Enum(fail), gl.Enum(zfail), gl.Enum(zpass))
}
func (gli GLImpl) TexImage2D(target uint32, level int32, internalformat int32, width int32, height int32, border int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(width * height * 4)
	sh.Len = int(width * height * 4)
	sh.Data = uintptr(pixels)
	gli.gl.TexImage2D(gl.Enum(target), int(level), int(internalformat), int(width), int(height), gl.Enum(format), gl.Enum(xtype), buf)
}
func (gli GLImpl) TexParameteri(target uint32, pname uint32, param int32) {
	gli.gl.TexParameteri(gl.Enum(target), gl.Enum(pname), int(param))
}
func (gli GLImpl) TexSubImage2D(target uint32, level int32, xoffset int32, yoffset int32, width int32, height int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(width * height * 4)
	sh.Len = int(width * height * 4)
	sh.Data = uintptr(pixels)
	gli.gl.TexSubImage2D(gl.Enum(target), int(level), int(xoffset), int(yoffset), int(width), int(height), gl.Enum(format), gl.Enum(xtype), buf)
}
func (gli GLImpl) Uniform1f(location int32, v0 float32) {
	gli.gl.Uniform1f(gl.Uniform{Value: location}, v0)
}
func (gli GLImpl) Uniform1fv(location int32, count int32, v *float32) {
	var buf []float32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = int(count)
	sh.Len = int(count)
	sh.Data = uintptr(unsafe.Pointer(v))
	gli.gl.Uniform1fv(gl.Uniform{Value: location}, buf)
}
func (gli GLImpl) Uniform1i(location int32, v0 int32) {
	gli.gl.Uniform1i(gl.Uniform{Value: location}, int(v0))
}
func (gli GLImpl) Uniform2f(location int32, v0 float32, v1 float32) {
	gli.gl.Uniform2f(gl.Uniform{Value: location}, v0, v1)
}
func (gli GLImpl) Uniform4f(location int32, v0 float32, v1 float32, v2 float32, v3 float32) {
	gli.gl.Uniform4f(gl.Uniform{Value: location}, v0, v1, v2, v3)
}
func (gli GLImpl) UniformMatrix3fv(location int32, count int32, transpose bool, value *float32) {
	var buf []float32
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = 9
	sh.Len = 9
	sh.Data = uintptr(unsafe.Pointer(value))
	gli.gl.UniformMatrix3fv(gl.Uniform{Value: location}, buf)
}
func (gli GLImpl) UseProgram(program uint32) {
	gli.gl.UseProgram(gli.programs[program])
}
func (gli GLImpl) VertexAttribPointer(index uint32, size int32, xtype uint32, normalized bool, stride int32, offset uint32) {
	gli.gl.VertexAttribPointer(gl.Attrib{Value: uint(index)}, int(size), gl.Enum(xtype), normalized, int(stride), int(offset))
}
func (gli GLImpl) Viewport(x int32, y int32, width int32, height int32) {
	gli.gl.Viewport(int(x), int(y), int(width), int(height))
}

package glandroidimpl

// #include <stdlib.h>
// #include <GLES2/gl2.h>
// #cgo android LDFLAGS: -lGLESv2
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/tfriedel6/canvas"
)

type GLImpl struct{}

var _ canvas.GL = GLImpl{}

func (_ GLImpl) Ptr(data interface{}) unsafe.Pointer {
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
func (_ GLImpl) ActiveTexture(texture uint32) {
	C.glActiveTexture(C.uint(texture))
}
func (_ GLImpl) AttachShader(program uint32, shader uint32) {
	C.glAttachShader(C.uint(program), C.uint(shader))
}
func (_ GLImpl) BindBuffer(target uint32, buffer uint32) {
	C.glBindBuffer(C.uint(target), C.uint(buffer))
}
func (_ GLImpl) BindTexture(target uint32, texture uint32) {
	C.glBindTexture(C.uint(target), C.uint(texture))
}
func (_ GLImpl) BlendFunc(sfactor uint32, dfactor uint32) {
	C.glBlendFunc(C.uint(sfactor), C.uint(dfactor))
}
func (_ GLImpl) BufferData(target uint32, size int, data unsafe.Pointer, usage uint32) {
	C.glBufferData(C.uint(target), C.long(size), data, C.uint(usage))
}
func (_ GLImpl) Clear(mask uint32) {
	C.glClear(C.uint(mask))
}
func (_ GLImpl) ColorMask(red bool, green bool, blue bool, alpha bool) {
	var r, g, b, a C.uchar
	if red {
		r = 1
	}
	if green {
		g = 1
	}
	if blue {
		b = 1
	}
	if alpha {
		a = 1
	}
	C.glColorMask(r, g, b, a)
}
func (_ GLImpl) CompileShader(shader uint32) {
	C.glCompileShader(C.uint(shader))
}
func (_ GLImpl) CreateProgram() uint32 {
	return uint32(C.glCreateProgram())
}
func (_ GLImpl) CreateShader(xtype uint32) uint32 {
	return uint32(C.glCreateShader(C.uint(xtype)))
}
func (_ GLImpl) DeleteShader(shader uint32) {
	C.glDeleteShader(C.uint(shader))
}
func (_ GLImpl) DeleteTextures(n int32, textures *uint32) {
	C.glDeleteTextures(C.int(n), (*C.uint)(textures))
}
func (_ GLImpl) DisableVertexAttribArray(index uint32) {
	C.glDisableVertexAttribArray(C.uint(index))
}
func (_ GLImpl) DrawArrays(mode uint32, first int32, count int32) {
	C.glDrawArrays(C.uint(mode), C.int(first), C.int(count))
}
func (_ GLImpl) Enable(cap uint32) {
	C.glEnable(C.uint(cap))
}
func (_ GLImpl) EnableVertexAttribArray(index uint32) {
	C.glEnableVertexAttribArray(C.uint(index))
}
func (_ GLImpl) GenBuffers(n int32, buffers *uint32) {
	C.glGenBuffers(C.int(n), (*C.uint)(buffers))
}
func (_ GLImpl) GenTextures(n int32, textures *uint32) {
	C.glGenTextures(C.int(n), (*C.uint)(textures))
}
func (_ GLImpl) GenerateMipmap(target uint32) {
	C.glGenerateMipmap(C.uint(target))
}
func (_ GLImpl) GetAttribLocation(program uint32, name string) int32 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return int32(C.glGetAttribLocation(C.uint(program), cname))
}
func (_ GLImpl) GetError() uint32 {
	return uint32(C.glGetError())
}
func (_ GLImpl) GetProgramInfoLog(program uint32) string {
	var length C.int
	C.glGetProgramiv(C.uint(program), C.GL_INFO_LOG_LENGTH, &length)
	if length == 0 {
		return ""
	}
	clog := C.CBytes(make([]byte, int(length)+1))
	defer C.free(clog)
	C.glGetProgramInfoLog(C.uint(program), C.int(length), nil, (*C.char)(clog))
	return string(C.GoBytes(clog, length))
}
func (_ GLImpl) GetProgramiv(program uint32, pname uint32, params *int32) {
	C.glGetProgramiv(C.uint(program), C.uint(pname), (*C.int)(params))
}
func (_ GLImpl) GetShaderInfoLog(program uint32) string {
	var length C.int
	C.glGetShaderiv(C.uint(program), C.GL_INFO_LOG_LENGTH, &length)
	if length == 0 {
		return ""
	}
	clog := C.CBytes(make([]byte, int(length)+1))
	defer C.free(clog)
	C.glGetShaderInfoLog(C.uint(program), C.int(length), nil, (*C.char)(clog))
	return string(C.GoBytes(clog, length))
}
func (_ GLImpl) GetShaderiv(shader uint32, pname uint32, params *int32) {
	C.glGetShaderiv(C.uint(shader), C.uint(pname), (*C.int)(params))

}
func (_ GLImpl) GetUniformLocation(program uint32, name string) int32 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return int32(C.glGetUniformLocation(C.uint(program), cname))
}
func (_ GLImpl) LinkProgram(program uint32) {
	C.glLinkProgram(C.uint(program))
}
func (_ GLImpl) ReadPixels(x int32, y int32, width int32, height int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	C.glReadPixels(C.int(x), C.int(y), C.int(width), C.int(height), C.uint(format), C.uint(xtype), pixels)
}
func (_ GLImpl) Scissor(x int32, y int32, width int32, height int32) {
	C.glScissor(C.int(x), C.int(y), C.int(width), C.int(height))
}
func (_ GLImpl) ShaderSource(shader uint32, source string) {
	csource := C.CString(source)
	defer C.free(unsafe.Pointer(csource))
	C.glShaderSource(C.uint(shader), 1, &csource, nil)
}
func (_ GLImpl) StencilFunc(xfunc uint32, ref int32, mask uint32) {
	C.glStencilFunc(C.uint(xfunc), C.int(ref), C.uint(mask))
}
func (_ GLImpl) StencilMask(mask uint32) {
	C.glStencilMask(C.uint(mask))
}
func (_ GLImpl) StencilOp(fail uint32, zfail uint32, zpass uint32) {
	C.glStencilOp(C.uint(fail), C.uint(zfail), C.uint(zpass))
}
func (_ GLImpl) TexImage2D(target uint32, level int32, internalformat int32, width int32, height int32, border int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	C.glTexImage2D(C.uint(target), C.int(level), C.int(internalformat), C.int(width), C.int(height), C.int(border), C.uint(format), C.uint(xtype), pixels)
}
func (_ GLImpl) TexParameteri(target uint32, pname uint32, param int32) {
	C.glTexParameteri(C.uint(target), C.uint(pname), C.int(param))
}
func (_ GLImpl) TexSubImage2D(target uint32, level int32, xoffset int32, yoffset int32, width int32, height int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	C.glTexSubImage2D(C.uint(target), C.int(level), C.int(xoffset), C.int(yoffset), C.int(width), C.int(height), C.uint(format), C.uint(xtype), pixels)
}
func (_ GLImpl) Uniform1f(location int32, v0 float32) {
	C.glUniform1f(C.int(location), C.float(v0))
}
func (_ GLImpl) Uniform1i(location int32, v0 int32) {
	C.glUniform1i(C.int(location), C.int(v0))
}
func (_ GLImpl) Uniform2f(location int32, v0 float32, v1 float32) {
	C.glUniform2f(C.int(location), C.float(v0), C.float(v1))
}
func (_ GLImpl) Uniform4f(location int32, v0 float32, v1 float32, v2 float32, v3 float32) {
	C.glUniform4f(C.int(location), C.float(v0), C.float(v1), C.float(v2), C.float(v3))
}
func (_ GLImpl) UniformMatrix3fv(location int32, count int32, transpose bool, value *float32) {
	var t C.uchar
	if transpose {
		t = 1
	}
	C.glUniformMatrix3fv(C.int(location), C.int(count), t, (*C.float)(value))
}
func (_ GLImpl) UseProgram(program uint32) {
	C.glUseProgram(C.uint(program))
}
func (_ GLImpl) VertexAttribPointer(index uint32, size int32, xtype uint32, normalized bool, stride int32, offset uint32) {
	var n C.uchar
	if normalized {
		n = 1
	}
	C.glVertexAttribPointer(C.uint(index), C.int(size), C.uint(xtype), n, C.int(stride), unsafe.Pointer(uintptr(offset)))
}
func (_ GLImpl) Viewport(x int32, y int32, width int32, height int32) {
	C.glViewport(C.int(x), C.int(y), C.int(width), C.int(height))
}

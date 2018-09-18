package glimplios

// #include <stdlib.h>
// #include <OpenGLES/ES2/gl.h>
// #cgo ios LDFLAGS: -framework OpenGLES
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
	C.glActiveTexture(C.GLenum(texture))
}
func (_ GLImpl) AttachShader(program uint32, shader uint32) {
	C.glAttachShader(C.GLuint(program), C.GLuint(shader))
}
func (_ GLImpl) BindBuffer(target uint32, buffer uint32) {
	C.glBindBuffer(C.GLenum(target), C.GLuint(buffer))
}
func (_ GLImpl) BindFramebuffer(target uint32, framebuffer uint32) {
	C.glBindFramebuffer(C.GLenum(target), C.GLuint(framebuffer))
}
func (_ GLImpl) BindRenderbuffer(target uint32, renderbuffer uint32) {
	C.glBindRenderbuffer(C.GLenum(target), C.GLuint(renderbuffer))
}
func (_ GLImpl) BindTexture(target uint32, texture uint32) {
	C.glBindTexture(C.GLenum(target), C.GLuint(texture))
}
func (_ GLImpl) BlendFunc(sfactor uint32, dfactor uint32) {
	C.glBlendFunc(C.GLenum(sfactor), C.GLenum(dfactor))
}
func (_ GLImpl) BufferData(target uint32, size int, data unsafe.Pointer, usage uint32) {
	C.glBufferData(C.GLenum(target), C.GLsizeiptr(size), data, C.GLenum(usage))
}
func (_ GLImpl) CheckFramebufferStatus(target uint32) uint32 {
	return uint32(C.glCheckFramebufferStatus(C.GLenum(target)))
}
func (_ GLImpl) Clear(mask uint32) {
	C.glClear(C.GLbitfield(mask))
}
func (_ GLImpl) ClearColor(red float32, green float32, blue float32, alpha float32) {
	C.glClearColor(C.GLfloat(red), C.GLfloat(green), C.GLfloat(blue), C.GLfloat(alpha))
}
func (_ GLImpl) ColorMask(red bool, green bool, blue bool, alpha bool) {
	var r, g, b, a C.GLboolean
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
	C.glCompileShader(C.GLuint(shader))
}
func (_ GLImpl) CreateProgram() uint32 {
	return uint32(C.glCreateProgram())
}
func (_ GLImpl) CreateShader(xtype uint32) uint32 {
	return uint32(C.glCreateShader(C.GLenum(xtype)))
}
func (_ GLImpl) DeleteShader(shader uint32) {
	C.glDeleteShader(C.GLuint(shader))
}
func (_ GLImpl) DeleteFramebuffers(n int32, framebuffers *uint32) {
	C.glDeleteFramebuffers(C.GLsizei(n), (*C.GLuint)(framebuffers))
}
func (_ GLImpl) DeleteRenderbuffers(n int32, renderbuffers *uint32) {
	C.glDeleteRenderbuffers(C.GLsizei(n), (*C.GLuint)(renderbuffers))
}
func (_ GLImpl) DeleteTextures(n int32, textures *uint32) {
	C.glDeleteTextures(C.GLsizei(n), (*C.GLuint)(textures))
}
func (_ GLImpl) Disable(cap uint32) {
	C.glDisable(C.GLenum(cap))
}
func (_ GLImpl) DisableVertexAttribArray(index uint32) {
	C.glDisableVertexAttribArray(C.GLuint(index))
}
func (_ GLImpl) DrawArrays(mode uint32, first int32, count int32) {
	C.glDrawArrays(C.GLenum(mode), C.GLint(first), C.GLsizei(count))
}
func (_ GLImpl) Enable(cap uint32) {
	C.glEnable(C.GLenum(cap))
}
func (_ GLImpl) EnableVertexAttribArray(index uint32) {
	C.glEnableVertexAttribArray(C.GLuint(index))
}
func (_ GLImpl) FramebufferRenderbuffer(target uint32, attachment uint32, renderbuffertarget uint32, renderbuffer uint32) {
	C.glFramebufferRenderbuffer(C.GLenum(target), C.GLenum(attachment), C.GLenum(renderbuffertarget), C.GLuint(renderbuffer))
}
func (_ GLImpl) FramebufferTexture(target uint32, attachment uint32, texture uint32, level int32) {
	C.glFramebufferTexture2D(C.GLenum(target), C.GLenum(attachment), C.GL_TEXTURE_2D, C.GLuint(texture), C.GLint(level))
}
func (_ GLImpl) GenBuffers(n int32, buffers *uint32) {
	C.glGenBuffers(C.GLsizei(n), (*C.GLuint)(buffers))
}
func (_ GLImpl) GenFramebuffers(n int32, framebuffers *uint32) {
	C.glGenFramebuffers(C.GLsizei(n), (*C.GLuint)(framebuffers))
}
func (_ GLImpl) GenRenderbuffers(n int32, renderbuffers *uint32) {
	C.glGenRenderbuffers(C.GLsizei(n), (*C.GLuint)(renderbuffers))
}
func (_ GLImpl) GenTextures(n int32, textures *uint32) {
	C.glGenTextures(C.GLsizei(n), (*C.GLuint)(textures))
}
func (_ GLImpl) GenerateMipmap(target uint32) {
	C.glGenerateMipmap(C.GLenum(target))
}
func (_ GLImpl) GetAttribLocation(program uint32, name string) int32 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return int32(C.glGetAttribLocation(C.GLuint(program), (*C.GLchar)(cname)))
}
func (_ GLImpl) GetError() uint32 {
	return uint32(C.glGetError())
}
func (_ GLImpl) GetProgramInfoLog(program uint32) string {
	var length C.GLint
	C.glGetProgramiv(C.GLuint(program), C.GL_INFO_LOG_LENGTH, &length)
	if length == 0 {
		return ""
	}
	clog := C.CBytes(make([]byte, int(length)+1))
	defer C.free(clog)
	C.glGetProgramInfoLog(C.GLuint(program), C.GLsizei(length), nil, (*C.GLchar)(clog))
	return string(C.GoBytes(clog, C.int(length)))
}
func (_ GLImpl) GetProgramiv(program uint32, pname uint32, params *int32) {
	C.glGetProgramiv(C.GLuint(program), C.GLenum(pname), (*C.GLint)(params))
}
func (_ GLImpl) GetShaderInfoLog(program uint32) string {
	var length C.GLint
	C.glGetShaderiv(C.GLuint(program), C.GL_INFO_LOG_LENGTH, &length)
	if length == 0 {
		return ""
	}
	clog := C.CBytes(make([]byte, int(length)+1))
	defer C.free(clog)
	C.glGetShaderInfoLog(C.GLuint(program), C.GLsizei(length), nil, (*C.GLchar)(clog))
	return string(C.GoBytes(clog, C.int(length)))
}
func (_ GLImpl) GetShaderiv(shader uint32, pname uint32, params *int32) {
	C.glGetShaderiv(C.GLuint(shader), C.GLenum(pname), (*C.GLint)(params))

}
func (_ GLImpl) GetUniformLocation(program uint32, name string) int32 {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return int32(C.glGetUniformLocation(C.GLuint(program), (*C.GLchar)(cname)))
}
func (_ GLImpl) LinkProgram(program uint32) {
	C.glLinkProgram(C.GLuint(program))
}
func (_ GLImpl) ReadPixels(x int32, y int32, width int32, height int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	C.glReadPixels(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(xtype), pixels)
}
func (_ GLImpl) RenderbufferStorage(target uint32, internalformat uint32, width int32, height int32) {
	C.glRenderbufferStorage(C.GLenum(target), C.GLenum(internalformat), C.GLint(width), C.GLint(height))
}
func (_ GLImpl) Scissor(x int32, y int32, width int32, height int32) {
	C.glScissor(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}
func (_ GLImpl) ShaderSource(shader uint32, source string) {
	csource := (*C.GLchar)(C.CString(source))
	defer C.free(unsafe.Pointer(csource))
	C.glShaderSource(C.GLuint(shader), 1, &csource, nil)
}
func (_ GLImpl) StencilFunc(xfunc uint32, ref int32, mask uint32) {
	C.glStencilFunc(C.GLenum(xfunc), C.GLint(ref), C.GLuint(mask))
}
func (_ GLImpl) StencilMask(mask uint32) {
	C.glStencilMask(C.GLuint(mask))
}
func (_ GLImpl) StencilOp(fail uint32, zfail uint32, zpass uint32) {
	C.glStencilOp(C.GLenum(fail), C.GLenum(zfail), C.GLenum(zpass))
}
func (_ GLImpl) TexImage2D(target uint32, level int32, internalformat int32, width int32, height int32, border int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	C.glTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalformat), C.GLsizei(width), C.GLsizei(height), C.GLint(border), C.GLenum(format), C.GLenum(xtype), pixels)
}
func (_ GLImpl) TexParameteri(target uint32, pname uint32, param int32) {
	C.glTexParameteri(C.GLenum(target), C.GLenum(pname), C.GLint(param))
}
func (_ GLImpl) TexSubImage2D(target uint32, level int32, xoffset int32, yoffset int32, width int32, height int32, format uint32, xtype uint32, pixels unsafe.Pointer) {
	C.glTexSubImage2D(C.GLenum(target), C.GLint(level), C.GLint(xoffset), C.GLint(yoffset), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(xtype), pixels)
}
func (_ GLImpl) Uniform1f(location int32, v0 float32) {
	C.glUniform1f(C.GLint(location), C.GLfloat(v0))
}
func (_ GLImpl) Uniform1fv(location int32, count int32, v *float32) {
	C.glUniform1fv(C.GLint(location), C.GLsizei(count), (*C.GLfloat)(v))
}
func (_ GLImpl) Uniform1i(location int32, v0 int32) {
	C.glUniform1i(C.GLint(location), C.GLint(v0))
}
func (_ GLImpl) Uniform2f(location int32, v0 float32, v1 float32) {
	C.glUniform2f(C.GLint(location), C.GLfloat(v0), C.GLfloat(v1))
}
func (_ GLImpl) Uniform4f(location int32, v0 float32, v1 float32, v2 float32, v3 float32) {
	C.glUniform4f(C.GLint(location), C.GLfloat(v0), C.GLfloat(v1), C.GLfloat(v2), C.GLfloat(v3))
}
func (_ GLImpl) UniformMatrix3fv(location int32, count int32, transpose bool, value *float32) {
	var t C.GLboolean
	if transpose {
		t = 1
	}
	C.glUniformMatrix3fv(C.GLint(location), C.GLsizei(count), t, (*C.GLfloat)(value))
}
func (_ GLImpl) UseProgram(program uint32) {
	C.glUseProgram(C.GLuint(program))
}
func (_ GLImpl) VertexAttribPointer(index uint32, size int32, xtype uint32, normalized bool, stride int32, offset uint32) {
	var n C.GLboolean
	if normalized {
		n = 1
	}
	C.glVertexAttribPointer(C.GLuint(index), C.GLint(size), C.GLenum(xtype), n, C.GLsizei(stride), unsafe.Pointer(uintptr(offset)))
}
func (_ GLImpl) Viewport(x int32, y int32, width int32, height int32) {
	C.glViewport(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

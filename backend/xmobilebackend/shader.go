package xmobilebackend

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/mobile/gl"
)

type shaderProgram struct {
	b      *XMobileBackend
	ID     gl.Program
	vs, fs gl.Shader

	attribs  map[string]gl.Attrib
	uniforms map[string]gl.Uniform
}

func loadShader(b *XMobileBackend, vs, fs string, sp *shaderProgram) error {
	sp.b = b
	glError(b) // clear the current error

	// compile vertex shader
	{
		sp.vs = b.glctx.CreateShader(gl.VERTEX_SHADER)
		b.glctx.ShaderSource(sp.vs, vs)
		b.glctx.CompileShader(sp.vs)

		status := b.glctx.GetShaderi(sp.vs, gl.COMPILE_STATUS)
		if status != gl.TRUE {
			clog := b.glctx.GetShaderInfoLog(sp.vs)
			b.glctx.DeleteShader(sp.vs)
			return fmt.Errorf("failed to compile vertex shader:\n\n%s", clog)
		}
		if err := glError(b); err != nil {
			return fmt.Errorf("gl error after compiling vertex shader: %v", err)
		}
	}

	// compile fragment shader
	{
		sp.fs = b.glctx.CreateShader(gl.FRAGMENT_SHADER)
		b.glctx.ShaderSource(sp.fs, fs)
		b.glctx.CompileShader(sp.fs)

		status := b.glctx.GetShaderi(sp.fs, gl.COMPILE_STATUS)
		if status != gl.TRUE {
			clog := b.glctx.GetShaderInfoLog(sp.fs)
			b.glctx.DeleteShader(sp.fs)
			return fmt.Errorf("failed to compile fragment shader:\n\n%s", clog)
		}
		if err := glError(b); err != nil {
			return fmt.Errorf("gl error after compiling fragment shader: %v", err)
		}
	}

	// link shader program
	{
		sp.ID = b.glctx.CreateProgram()
		b.glctx.AttachShader(sp.ID, sp.vs)
		b.glctx.AttachShader(sp.ID, sp.fs)
		b.glctx.LinkProgram(sp.ID)

		status := b.glctx.GetProgrami(sp.ID, gl.LINK_STATUS)
		if status != gl.TRUE {
			clog := b.glctx.GetProgramInfoLog(sp.ID)
			b.glctx.DeleteProgram(sp.ID)
			b.glctx.DeleteShader(sp.vs)
			b.glctx.DeleteShader(sp.fs)
			return fmt.Errorf("failed to link shader program:\n\n%s", clog)
		}
		if err := glError(b); err != nil {
			return fmt.Errorf("gl error after linking shader: %v", err)
		}
	}

	b.glctx.UseProgram(sp.ID)
	// load the attributes
	count := b.glctx.GetProgrami(sp.ID, gl.ACTIVE_ATTRIBUTES)
	sp.attribs = make(map[string]gl.Attrib, int(count))
	for i := 0; i < count; i++ {
		name, _, _ := b.glctx.GetActiveAttrib(sp.ID, uint32(i))
		sp.attribs[name] = b.glctx.GetAttribLocation(sp.ID, name)
	}

	// load the uniforms
	count = b.glctx.GetProgrami(sp.ID, gl.ACTIVE_UNIFORMS)
	sp.uniforms = make(map[string]gl.Uniform, int(count))
	for i := 0; i < count; i++ {
		name, _, _ := b.glctx.GetActiveUniform(sp.ID, uint32(i))
		sp.uniforms[name] = b.glctx.GetUniformLocation(sp.ID, name)
	}

	return nil
}

func (sp *shaderProgram) use() {
	sp.b.glctx.UseProgram(sp.ID)
}

func (sp *shaderProgram) delete() {
	sp.b.glctx.DeleteProgram(sp.ID)
	sp.b.glctx.DeleteShader(sp.vs)
	sp.b.glctx.DeleteShader(sp.fs)
}

func (sp *shaderProgram) loadLocations(target interface{}) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr {
		panic("target must be a pointer to a struct")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		panic("target must be a pointer to a struct")
	}

	sp.b.glctx.UseProgram(sp.ID)

	var errs strings.Builder

	for name, loc := range sp.attribs {
		field := val.FieldByName(sp.structName(name))
		if field == (reflect.Value{}) {
			fmt.Fprintf(&errs, "field for attribute \"%s\" not found; ", name)
		} else if field.Type() != reflect.TypeOf(gl.Attrib{}) {
			fmt.Fprintf(&errs, "field for attribute \"%s\" must have type gl.Attrib; ", name)
		} else {
			field.Set(reflect.ValueOf(loc))
		}
	}

	for name, loc := range sp.uniforms {
		field := val.FieldByName(sp.structName(name))
		if field == (reflect.Value{}) {
			fmt.Fprintf(&errs, "field for uniform \"%s\" not found; ", name)
		} else if field.Type() != reflect.TypeOf(gl.Uniform{}) {
			fmt.Fprintf(&errs, "field for uniform \"%s\" must have type gl.Uniform; ", name)
		} else {
			field.Set(reflect.ValueOf(loc))
		}
	}

	if errs.Len() > 0 {
		return errors.New(strings.TrimSpace(errs.String()))
	}
	return nil
}

func (sp *shaderProgram) structName(name string) string {
	rn, sz := utf8.DecodeRuneInString(name)
	name = fmt.Sprintf("%c%s", unicode.ToUpper(rn), name[sz:])
	idx := strings.IndexByte(name, '[')
	if idx > 0 {
		name = name[:idx]
	}
	return name
}

func (sp *shaderProgram) mustLoadLocations(target interface{}) {
	err := sp.loadLocations(target)
	if err != nil {
		panic(err)
	}
}

func (sp *shaderProgram) enableAllVertexAttribArrays() {
	for _, loc := range sp.attribs {
		sp.b.glctx.EnableVertexAttribArray(loc)
	}
}

func (sp *shaderProgram) disableAllVertexAttribArrays() {
	for _, loc := range sp.attribs {
		sp.b.glctx.DisableVertexAttribArray(loc)
	}
}

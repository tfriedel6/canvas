package goglbackend

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
)

type shaderProgram struct {
	ID, vs, fs uint32

	attribs  map[string]uint32
	uniforms map[string]int32
}

func loadShader(vs, fs string, sp *shaderProgram) error {
	glError() // clear the current error

	// compile vertex shader
	{
		sp.vs = gl.CreateShader(gl.VERTEX_SHADER)
		csrc, freeFunc := gl.Strs(vs + "\x00")
		defer freeFunc()
		gl.ShaderSource(sp.vs, 1, csrc, nil)
		gl.CompileShader(sp.vs)

		var status int32
		gl.GetShaderiv(sp.vs, gl.COMPILE_STATUS, &status)
		if status != gl.TRUE {
			var buf [65536]byte
			var length int32
			gl.GetShaderInfoLog(sp.vs, int32(len(buf)), &length, &buf[0])
			clog := string(buf[:length])
			gl.DeleteShader(sp.vs)
			return fmt.Errorf("failed to compile vertex shader:\n\n%s", clog)
		}
		if err := glError(); err != nil {
			return fmt.Errorf("gl error after compiling vertex shader: %v", err)
		}
	}

	// compile fragment shader
	{
		sp.fs = gl.CreateShader(gl.FRAGMENT_SHADER)
		csrc, freeFunc := gl.Strs(fs + "\x00")
		defer freeFunc()
		gl.ShaderSource(sp.fs, 1, csrc, nil)
		gl.CompileShader(sp.fs)

		var status int32
		gl.GetShaderiv(sp.fs, gl.COMPILE_STATUS, &status)
		if status != gl.TRUE {
			var buf [65536]byte
			var length int32
			gl.GetShaderInfoLog(sp.fs, int32(len(buf)), &length, &buf[0])
			clog := string(buf[:length])
			gl.DeleteShader(sp.fs)
			return fmt.Errorf("failed to compile fragment shader:\n\n%s", clog)
		}
		if err := glError(); err != nil {
			return fmt.Errorf("gl error after compiling fragment shader: %v", err)
		}
	}

	// link shader program
	{
		sp.ID = gl.CreateProgram()
		gl.AttachShader(sp.ID, sp.vs)
		gl.AttachShader(sp.ID, sp.fs)
		gl.LinkProgram(sp.ID)

		var status int32
		gl.GetProgramiv(sp.ID, gl.LINK_STATUS, &status)
		if status != gl.TRUE {
			var buf [65536]byte
			var length int32
			gl.GetProgramInfoLog(sp.ID, int32(len(buf)), &length, &buf[0])
			clog := string(buf[:length])
			gl.DeleteProgram(sp.ID)
			gl.DeleteShader(sp.vs)
			gl.DeleteShader(sp.fs)
			return fmt.Errorf("failed to link shader program:\n\n%s", clog)
		}
		if err := glError(); err != nil {
			return fmt.Errorf("gl error after linking shader: %v", err)
		}
	}

	gl.UseProgram(sp.ID)
	var nameBuf [256]byte
	var length, size int32
	var xtype uint32
	var count int32

	// load the attributes
	gl.GetProgramiv(sp.ID, gl.ACTIVE_ATTRIBUTES, &count)
	sp.attribs = make(map[string]uint32, int(count))
	for i := int32(0); i < count; i++ {
		gl.GetActiveAttrib(sp.ID, uint32(i), int32(len(nameBuf)), &length, &size, &xtype, &nameBuf[0])
		name := string(nameBuf[:length])
		loc := gl.GetAttribLocation(sp.ID, &nameBuf[0])
		sp.attribs[name] = uint32(loc)
	}

	// load the uniforms
	gl.GetProgramiv(sp.ID, gl.ACTIVE_UNIFORMS, &count)
	sp.uniforms = make(map[string]int32, int(count))
	for i := int32(0); i < count; i++ {
		gl.GetActiveUniform(sp.ID, uint32(i), int32(len(nameBuf)), &length, &size, &xtype, &nameBuf[0])
		name := string(nameBuf[:length])
		loc := gl.GetUniformLocation(sp.ID, &nameBuf[0])
		sp.uniforms[name] = loc
	}

	return nil
}

func (sp *shaderProgram) use() {
	gl.UseProgram(sp.ID)
}

func (sp *shaderProgram) delete() {
	gl.DeleteProgram(sp.ID)
	gl.DeleteShader(sp.vs)
	gl.DeleteShader(sp.fs)
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

	gl.UseProgram(sp.ID)

	var errs strings.Builder

	for name, loc := range sp.attribs {
		field := val.FieldByName(sp.structName(name))
		if field == (reflect.Value{}) {
			fmt.Fprintf(&errs, "field for attribute \"%s\" not found; ", name)
		} else if field.Type() != reflect.TypeOf(uint32(0)) {
			fmt.Fprintf(&errs, "field for attribute \"%s\" must have type uint32; ", name)
		} else {
			field.Set(reflect.ValueOf(uint32(loc)))
		}
	}

	for name, loc := range sp.uniforms {
		field := val.FieldByName(sp.structName(name))
		if field == (reflect.Value{}) {
			fmt.Fprintf(&errs, "field for uniform \"%s\" not found; ", name)
		} else if field.Type() != reflect.TypeOf(int32(0)) {
			fmt.Fprintf(&errs, "field for uniform \"%s\" must have type int32; ", name)
		} else {
			field.Set(reflect.ValueOf(int32(loc)))
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
		gl.EnableVertexAttribArray(loc)
	}
}

func (sp *shaderProgram) disableAllVertexAttribArrays() {
	for _, loc := range sp.attribs {
		gl.DisableVertexAttribArray(loc)
	}
}

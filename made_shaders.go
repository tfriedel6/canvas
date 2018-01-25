package canvas

import (
	"errors"
	"fmt"
	"strings"
)

type solidShader struct {
	id     uint32
	vertex uint32
	color  int32
}

func loadSolidShader() (*solidShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(solidVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(solidFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		program = gli.CreateProgram()
		gli.AttachShader(program, vs)
		gli.AttachShader(program, fs)
		gli.LinkProgram(program)

		var logLength int32
		gli.GetProgramiv(program, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetProgramInfoLog(program, logLength, nil, gli.Str(shLog))
			fmt.Printf("Shader link log:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &solidShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.color = gli.GetUniformLocation(program, gli.Str("color\x00"))

	return result, nil
}

type textureShader struct {
	id       uint32
	vertex   uint32
	texCoord uint32
	image    int32
}

func loadTextureShader() (*textureShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(textureVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(textureFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		program = gli.CreateProgram()
		gli.AttachShader(program, vs)
		gli.AttachShader(program, fs)
		gli.LinkProgram(program)

		var logLength int32
		gli.GetProgramiv(program, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetProgramInfoLog(program, logLength, nil, gli.Str(shLog))
			fmt.Printf("Shader link log:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &textureShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.texCoord = uint32(gli.GetAttribLocation(program, gli.Str("texCoord\x00")))
	result.image = gli.GetUniformLocation(program, gli.Str("image\x00"))

	return result, nil
}

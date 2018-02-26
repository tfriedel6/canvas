package canvas

import (
	"errors"
	"fmt"
	"strings"
)

type textureShader struct {
	id         uint32
	vertex     uint32
	texCoord   uint32
	canvasSize int32
	image      int32
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
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.image = gli.GetUniformLocation(program, gli.Str("image\x00"))

	return result, nil
}

type solidShader struct {
	id         uint32
	vertex     uint32
	canvasSize int32
	color      int32
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
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.color = gli.GetUniformLocation(program, gli.Str("color\x00"))

	return result, nil
}

type imagePatternShader struct {
	id         uint32
	vertex     uint32
	canvasSize int32
	imageSize  int32
	invmat     int32
	image      int32
}

func loadImagePatternShader() (*imagePatternShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(imagePatternVS + "\x00")
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
		csource, freeFunc := gli.Strs(imagePatternFS + "\x00")
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

	result := &imagePatternShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.imageSize = gli.GetUniformLocation(program, gli.Str("imageSize\x00"))
	result.invmat = gli.GetUniformLocation(program, gli.Str("invmat\x00"))
	result.image = gli.GetUniformLocation(program, gli.Str("image\x00"))

	return result, nil
}

type linearGradientShader struct {
	id         uint32
	vertex     uint32
	canvasSize int32
	invmat     int32
	gradient   int32
	from       int32
	dir        int32
	len        int32
}

func loadLinearGradientShader() (*linearGradientShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(linearGradientVS + "\x00")
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
		csource, freeFunc := gli.Strs(linearGradientFS + "\x00")
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

	result := &linearGradientShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.invmat = gli.GetUniformLocation(program, gli.Str("invmat\x00"))
	result.gradient = gli.GetUniformLocation(program, gli.Str("gradient\x00"))
	result.from = gli.GetUniformLocation(program, gli.Str("from\x00"))
	result.dir = gli.GetUniformLocation(program, gli.Str("dir\x00"))
	result.len = gli.GetUniformLocation(program, gli.Str("len\x00"))

	return result, nil
}

type radialGradientShader struct {
	id         uint32
	vertex     uint32
	canvasSize int32
	invmat     int32
	gradient   int32
	from       int32
	to         int32
	dir        int32
	radFrom    int32
	radTo      int32
	len        int32
}

func loadRadialGradientShader() (*radialGradientShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(radialGradientVS + "\x00")
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
		csource, freeFunc := gli.Strs(radialGradientFS + "\x00")
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

	result := &radialGradientShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.invmat = gli.GetUniformLocation(program, gli.Str("invmat\x00"))
	result.gradient = gli.GetUniformLocation(program, gli.Str("gradient\x00"))
	result.from = gli.GetUniformLocation(program, gli.Str("from\x00"))
	result.to = gli.GetUniformLocation(program, gli.Str("to\x00"))
	result.dir = gli.GetUniformLocation(program, gli.Str("dir\x00"))
	result.radFrom = gli.GetUniformLocation(program, gli.Str("radFrom\x00"))
	result.radTo = gli.GetUniformLocation(program, gli.Str("radTo\x00"))
	result.len = gli.GetUniformLocation(program, gli.Str("len\x00"))

	return result, nil
}

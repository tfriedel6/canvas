package canvas

import (
	"errors"
	"fmt"
	"strings"
)

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
			fmt.Printf("VERTEX_SHADER compilation log for linearGradientVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for linearGradientVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for linearGradientVS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("FRAGMENT_SHADER compilation log for linearGradientFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for linearGradientFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for linearGradientFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for linearGradientFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for linearGradientFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for linearGradientFS, glError: " + fmt.Sprint(glErr))
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

type linearGradientAlphaShader struct {
	id            uint32
	vertex        uint32
	alphaTexCoord uint32
	canvasSize    int32
	invmat        int32
	gradient      int32
	from          int32
	dir           int32
	len           int32
	alphaTex      int32
}

func loadLinearGradientAlphaShader() (*linearGradientAlphaShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(linearGradientAlphaVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log for linearGradientAlphaVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for linearGradientAlphaVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for linearGradientAlphaVS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(linearGradientAlphaFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log for linearGradientAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for linearGradientAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for linearGradientAlphaFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for linearGradientAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for linearGradientAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for linearGradientAlphaFS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &linearGradientAlphaShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, gli.Str("alphaTexCoord\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.invmat = gli.GetUniformLocation(program, gli.Str("invmat\x00"))
	result.gradient = gli.GetUniformLocation(program, gli.Str("gradient\x00"))
	result.from = gli.GetUniformLocation(program, gli.Str("from\x00"))
	result.dir = gli.GetUniformLocation(program, gli.Str("dir\x00"))
	result.len = gli.GetUniformLocation(program, gli.Str("len\x00"))
	result.alphaTex = gli.GetUniformLocation(program, gli.Str("alphaTex\x00"))

	return result, nil
}

type imagePatternAlphaShader struct {
	id            uint32
	vertex        uint32
	alphaTexCoord uint32
	canvasSize    int32
	imageSize     int32
	invmat        int32
	image         int32
	alphaTex      int32
}

func loadImagePatternAlphaShader() (*imagePatternAlphaShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(imagePatternAlphaVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log for imagePatternAlphaVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for imagePatternAlphaVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for imagePatternAlphaVS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(imagePatternAlphaFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log for imagePatternAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for imagePatternAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for imagePatternAlphaFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for imagePatternAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for imagePatternAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for imagePatternAlphaFS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &imagePatternAlphaShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, gli.Str("alphaTexCoord\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.imageSize = gli.GetUniformLocation(program, gli.Str("imageSize\x00"))
	result.invmat = gli.GetUniformLocation(program, gli.Str("invmat\x00"))
	result.image = gli.GetUniformLocation(program, gli.Str("image\x00"))
	result.alphaTex = gli.GetUniformLocation(program, gli.Str("alphaTex\x00"))

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
			fmt.Printf("VERTEX_SHADER compilation log for imagePatternVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for imagePatternVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for imagePatternVS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("FRAGMENT_SHADER compilation log for imagePatternFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for imagePatternFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for imagePatternFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for imagePatternFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for imagePatternFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for imagePatternFS, glError: " + fmt.Sprint(glErr))
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

type solidAlphaShader struct {
	id            uint32
	vertex        uint32
	alphaTexCoord uint32
	canvasSize    int32
	color         int32
	alphaTex      int32
}

func loadSolidAlphaShader() (*solidAlphaShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(solidAlphaVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log for solidAlphaVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for solidAlphaVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for solidAlphaVS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(solidAlphaFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log for solidAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for solidAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for solidAlphaFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for solidAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for solidAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for solidAlphaFS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &solidAlphaShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, gli.Str("alphaTexCoord\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.color = gli.GetUniformLocation(program, gli.Str("color\x00"))
	result.alphaTex = gli.GetUniformLocation(program, gli.Str("alphaTex\x00"))

	return result, nil
}

type radialGradientAlphaShader struct {
	id            uint32
	vertex        uint32
	alphaTexCoord uint32
	canvasSize    int32
	invmat        int32
	gradient      int32
	from          int32
	to            int32
	dir           int32
	radFrom       int32
	radTo         int32
	len           int32
	alphaTex      int32
}

func loadRadialGradientAlphaShader() (*radialGradientAlphaShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(radialGradientAlphaVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log for radialGradientAlphaVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for radialGradientAlphaVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for radialGradientAlphaVS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(radialGradientAlphaFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log for radialGradientAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for radialGradientAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for radialGradientAlphaFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for radialGradientAlphaFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for radialGradientAlphaFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for radialGradientAlphaFS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &radialGradientAlphaShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, gli.Str("alphaTexCoord\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.invmat = gli.GetUniformLocation(program, gli.Str("invmat\x00"))
	result.gradient = gli.GetUniformLocation(program, gli.Str("gradient\x00"))
	result.from = gli.GetUniformLocation(program, gli.Str("from\x00"))
	result.to = gli.GetUniformLocation(program, gli.Str("to\x00"))
	result.dir = gli.GetUniformLocation(program, gli.Str("dir\x00"))
	result.radFrom = gli.GetUniformLocation(program, gli.Str("radFrom\x00"))
	result.radTo = gli.GetUniformLocation(program, gli.Str("radTo\x00"))
	result.len = gli.GetUniformLocation(program, gli.Str("len\x00"))
	result.alphaTex = gli.GetUniformLocation(program, gli.Str("alphaTex\x00"))

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
			fmt.Printf("VERTEX_SHADER compilation log for solidVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for solidVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for solidVS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("FRAGMENT_SHADER compilation log for solidFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for solidFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for solidFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for solidFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for solidFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for solidFS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &solidShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.color = gli.GetUniformLocation(program, gli.Str("color\x00"))

	return result, nil
}

type imageShader struct {
	id         uint32
	vertex     uint32
	texCoord   uint32
	canvasSize int32
	image      int32
}

func loadImageShader() (*imageShader, error) {
	var vs, fs, program uint32

	{
		csource, freeFunc := gli.Strs(imageVS + "\x00")
		defer freeFunc()

		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, 1, csource, nil)
		gli.CompileShader(vs)

		var logLength int32
		gli.GetShaderiv(vs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(vs, logLength, nil, gli.Str(shLog))
			fmt.Printf("VERTEX_SHADER compilation log for imageVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for imageVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for imageVS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		csource, freeFunc := gli.Strs(imageFS + "\x00")
		defer freeFunc()

		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, 1, csource, nil)
		gli.CompileShader(fs)

		var logLength int32
		gli.GetShaderiv(fs, gl_INFO_LOG_LENGTH, &logLength)
		if logLength > 0 {
			shLog := strings.Repeat("\x00", int(logLength+1))
			gli.GetShaderInfoLog(fs, logLength, nil, gli.Str(shLog))
			fmt.Printf("FRAGMENT_SHADER compilation log for imageFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for imageFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for imageFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for imageFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for imageFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for imageFS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &imageShader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, gli.Str("vertex\x00")))
	result.texCoord = uint32(gli.GetAttribLocation(program, gli.Str("texCoord\x00")))
	result.canvasSize = gli.GetUniformLocation(program, gli.Str("canvasSize\x00"))
	result.image = gli.GetUniformLocation(program, gli.Str("image\x00"))

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
			fmt.Printf("VERTEX_SHADER compilation log for radialGradientVS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for radialGradientVS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for radialGradientVS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("FRAGMENT_SHADER compilation log for radialGradientFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for radialGradientFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for radialGradientFS, glError: " + fmt.Sprint(glErr))
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
			fmt.Printf("Shader link log for radialGradientFS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for radialGradientFS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for radialGradientFS, glError: " + fmt.Sprint(glErr))
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

package canvas

import (
	"errors"
	"fmt"
)

type gaussian63Shader struct {
	id          uint32
	vertex      uint32
	texCoord    uint32
	canvasSize  int32
	kernelScale int32
	image       int32
	kernel      int32
}

func loadGaussian63Shader() (*gaussian63Shader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, gaussian63VS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
			fmt.Printf("VERTEX_SHADER compilation log for gaussian63VS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for gaussian63VS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for gaussian63VS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, gaussian63FS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
			fmt.Printf("FRAGMENT_SHADER compilation log for gaussian63FS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for gaussian63FS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for gaussian63FS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		program = gli.CreateProgram()
		gli.AttachShader(program, vs)
		gli.AttachShader(program, fs)
		gli.LinkProgram(program)

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
			fmt.Printf("Shader link log for gaussian63FS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for gaussian63FS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for gaussian63FS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &gaussian63Shader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.texCoord = uint32(gli.GetAttribLocation(program, "texCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.kernelScale = gli.GetUniformLocation(program, "kernelScale")
	result.image = gli.GetUniformLocation(program, "image")
	result.kernel = gli.GetUniformLocation(program, "kernel")

	return result, nil
}

type solidShader struct {
	id          uint32
	vertex      uint32
	canvasSize  int32
	color       int32
	globalAlpha int32
}

func loadSolidShader() (*solidShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, solidVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, solidFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.color = gli.GetUniformLocation(program, "color")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

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
	globalAlpha   int32
}

func loadLinearGradientAlphaShader() (*linearGradientAlphaShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, linearGradientAlphaVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, linearGradientAlphaFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, "alphaTexCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.invmat = gli.GetUniformLocation(program, "invmat")
	result.gradient = gli.GetUniformLocation(program, "gradient")
	result.from = gli.GetUniformLocation(program, "from")
	result.dir = gli.GetUniformLocation(program, "dir")
	result.len = gli.GetUniformLocation(program, "len")
	result.alphaTex = gli.GetUniformLocation(program, "alphaTex")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

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
	globalAlpha   int32
}

func loadRadialGradientAlphaShader() (*radialGradientAlphaShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, radialGradientAlphaVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, radialGradientAlphaFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, "alphaTexCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.invmat = gli.GetUniformLocation(program, "invmat")
	result.gradient = gli.GetUniformLocation(program, "gradient")
	result.from = gli.GetUniformLocation(program, "from")
	result.to = gli.GetUniformLocation(program, "to")
	result.dir = gli.GetUniformLocation(program, "dir")
	result.radFrom = gli.GetUniformLocation(program, "radFrom")
	result.radTo = gli.GetUniformLocation(program, "radTo")
	result.len = gli.GetUniformLocation(program, "len")
	result.alphaTex = gli.GetUniformLocation(program, "alphaTex")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

	return result, nil
}

type gaussian15Shader struct {
	id          uint32
	vertex      uint32
	texCoord    uint32
	canvasSize  int32
	kernelScale int32
	image       int32
	kernel      int32
}

func loadGaussian15Shader() (*gaussian15Shader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, gaussian15VS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
			fmt.Printf("VERTEX_SHADER compilation log for gaussian15VS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for gaussian15VS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for gaussian15VS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, gaussian15FS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
			fmt.Printf("FRAGMENT_SHADER compilation log for gaussian15FS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for gaussian15FS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for gaussian15FS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		program = gli.CreateProgram()
		gli.AttachShader(program, vs)
		gli.AttachShader(program, fs)
		gli.LinkProgram(program)

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
			fmt.Printf("Shader link log for gaussian15FS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for gaussian15FS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for gaussian15FS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &gaussian15Shader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.texCoord = uint32(gli.GetAttribLocation(program, "texCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.kernelScale = gli.GetUniformLocation(program, "kernelScale")
	result.image = gli.GetUniformLocation(program, "image")
	result.kernel = gli.GetUniformLocation(program, "kernel")

	return result, nil
}

type imagePatternShader struct {
	id          uint32
	vertex      uint32
	canvasSize  int32
	imageSize   int32
	invmat      int32
	image       int32
	globalAlpha int32
}

func loadImagePatternShader() (*imagePatternShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, imagePatternVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, imagePatternFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.imageSize = gli.GetUniformLocation(program, "imageSize")
	result.invmat = gli.GetUniformLocation(program, "invmat")
	result.image = gli.GetUniformLocation(program, "image")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

	return result, nil
}

type solidAlphaShader struct {
	id            uint32
	vertex        uint32
	alphaTexCoord uint32
	canvasSize    int32
	color         int32
	alphaTex      int32
	globalAlpha   int32
}

func loadSolidAlphaShader() (*solidAlphaShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, solidAlphaVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, solidAlphaFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, "alphaTexCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.color = gli.GetUniformLocation(program, "color")
	result.alphaTex = gli.GetUniformLocation(program, "alphaTex")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

	return result, nil
}

type gaussian255Shader struct {
	id          uint32
	vertex      uint32
	texCoord    uint32
	canvasSize  int32
	kernelScale int32
	image       int32
	kernel      int32
}

func loadGaussian255Shader() (*gaussian255Shader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, gaussian255VS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
			fmt.Printf("VERTEX_SHADER compilation log for gaussian255VS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(vs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			return nil, errors.New("Error compiling GL_VERTEX_SHADER shader part for gaussian255VS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for gaussian255VS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, gaussian255FS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
			fmt.Printf("FRAGMENT_SHADER compilation log for gaussian255FS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(fs, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(fs)
			return nil, errors.New("Error compiling GL_FRAGMENT_SHADER shader part for gaussian255FS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for gaussian255FS, glError: " + fmt.Sprint(glErr))
		}
	}

	{
		program = gli.CreateProgram()
		gli.AttachShader(program, vs)
		gli.AttachShader(program, fs)
		gli.LinkProgram(program)

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
			fmt.Printf("Shader link log for gaussian255FS:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for gaussian255FS")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for gaussian255FS, glError: " + fmt.Sprint(glErr))
		}
	}

	result := &gaussian255Shader{}
	result.id = program
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.texCoord = uint32(gli.GetAttribLocation(program, "texCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.kernelScale = gli.GetUniformLocation(program, "kernelScale")
	result.image = gli.GetUniformLocation(program, "image")
	result.kernel = gli.GetUniformLocation(program, "kernel")

	return result, nil
}

type radialGradientShader struct {
	id          uint32
	vertex      uint32
	canvasSize  int32
	invmat      int32
	gradient    int32
	from        int32
	to          int32
	dir         int32
	radFrom     int32
	radTo       int32
	len         int32
	globalAlpha int32
}

func loadRadialGradientShader() (*radialGradientShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, radialGradientVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, radialGradientFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.invmat = gli.GetUniformLocation(program, "invmat")
	result.gradient = gli.GetUniformLocation(program, "gradient")
	result.from = gli.GetUniformLocation(program, "from")
	result.to = gli.GetUniformLocation(program, "to")
	result.dir = gli.GetUniformLocation(program, "dir")
	result.radFrom = gli.GetUniformLocation(program, "radFrom")
	result.radTo = gli.GetUniformLocation(program, "radTo")
	result.len = gli.GetUniformLocation(program, "len")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

	return result, nil
}

type linearGradientShader struct {
	id          uint32
	vertex      uint32
	canvasSize  int32
	invmat      int32
	gradient    int32
	from        int32
	dir         int32
	len         int32
	globalAlpha int32
}

func loadLinearGradientShader() (*linearGradientShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, linearGradientVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, linearGradientFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.invmat = gli.GetUniformLocation(program, "invmat")
	result.gradient = gli.GetUniformLocation(program, "gradient")
	result.from = gli.GetUniformLocation(program, "from")
	result.dir = gli.GetUniformLocation(program, "dir")
	result.len = gli.GetUniformLocation(program, "len")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

	return result, nil
}

type imageShader struct {
	id          uint32
	vertex      uint32
	texCoord    uint32
	canvasSize  int32
	image       int32
	globalAlpha int32
}

func loadImageShader() (*imageShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, imageVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, imageFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.texCoord = uint32(gli.GetAttribLocation(program, "texCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.image = gli.GetUniformLocation(program, "image")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

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
	globalAlpha   int32
}

func loadImagePatternAlphaShader() (*imagePatternAlphaShader, error) {
	var vs, fs, program uint32

	{
		vs = gli.CreateShader(gl_VERTEX_SHADER)
		gli.ShaderSource(vs, imagePatternAlphaVS)
		gli.CompileShader(vs)

		shLog := gli.GetShaderInfoLog(vs)
		if len(shLog) > 0 {
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
		fs = gli.CreateShader(gl_FRAGMENT_SHADER)
		gli.ShaderSource(fs, imagePatternAlphaFS)
		gli.CompileShader(fs)

		shLog := gli.GetShaderInfoLog(fs)
		if len(shLog) > 0 {
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

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
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
	result.vertex = uint32(gli.GetAttribLocation(program, "vertex"))
	result.alphaTexCoord = uint32(gli.GetAttribLocation(program, "alphaTexCoord"))
	result.canvasSize = gli.GetUniformLocation(program, "canvasSize")
	result.imageSize = gli.GetUniformLocation(program, "imageSize")
	result.invmat = gli.GetUniformLocation(program, "invmat")
	result.image = gli.GetUniformLocation(program, "image")
	result.alphaTex = gli.GetUniformLocation(program, "alphaTex")
	result.globalAlpha = gli.GetUniformLocation(program, "globalAlpha")

	return result, nil
}

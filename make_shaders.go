//+build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"unicode"
)

func main() {
	// find the go files in the current directory
	pkg, err := build.Default.ImportDir(".", 0)
	if err != nil {
		log.Fatalf("Could not process directory: %s", err)
	}

	vsMap := make(map[string]string)
	fsMap := make(map[string]string)

	// go through each file and find const raw string literals with
	fset := token.NewFileSet()
	for _, goFile := range pkg.GoFiles {
		parsedFile, err := parser.ParseFile(fset, goFile, nil, 0)
		if err != nil {
			log.Fatalf("Failed to parse file %s: %s", goFile, err)
		}

		for _, obj := range parsedFile.Scope.Objects {
			isVS := strings.HasSuffix(obj.Name, "VS")
			isFS := strings.HasSuffix(obj.Name, "FS")
			var shaderCode string
			if isVS || isFS {
				if value, ok := obj.Decl.(*ast.ValueSpec); ok && len(value.Values) == 1 {
					if lit, ok := value.Values[0].(*ast.BasicLit); ok {
						shaderCode = lit.Value
					}
				}
			}
			if shaderCode != "" {
				baseName := obj.Name[:len(obj.Name)-2]
				if isVS {
					vsMap[baseName] = shaderCode
				} else if isFS {
					fsMap[baseName] = shaderCode
				}
			}
		}
	}

	for name := range vsMap {
		if _, ok := fsMap[name]; !ok {
			log.Println("Warning: Vertex shader with no corresponding fragment shader (" + name + ")")
			delete(vsMap, name)
			continue
		}
	}
	for name := range fsMap {
		if _, ok := vsMap[name]; !ok {
			log.Println("Warning: Fragment shader with no corresponding vertex shader (" + name + ")")
			delete(fsMap, name)
			continue
		}
	}

	goCode := &bytes.Buffer{}
	buildCodeHeader(goCode)

	for name, vs := range vsMap {
		fs := fsMap[name]

		vs = vs[1 : len(vs)-1]
		fs = fs[1 : len(fs)-1]

		inputs := shaderFindInputVariables(vs + "\n" + fs)

		buildCode(goCode, name, inputs)
	}

	err = ioutil.WriteFile("made_shaders.go", goCode.Bytes(), 0777)
	if err != nil {
		log.Fatalf("Could not write made_shaders.go: %s", err)
	}
}

type ShaderInput struct {
	Name      string
	IsAttrib  bool
	IsUniform bool
}

func shaderFindInputVariables(source string) []ShaderInput {
	inputs := make([]ShaderInput, 0, 10)

	varDefSplitter := regexp.MustCompile("[ \t\r\n,]+")
	lines := shaderGetTopLevelLines(source)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if strings.Contains(line, "{") {
			break
		}
		parts := varDefSplitter.Split(line, -1)
		isAttrib := parts[0] == "attribute"
		isUniform := parts[0] == "uniform"
		if !isAttrib && !isUniform {
			continue
		}
		for _, part := range parts[2:] {
			if idx := strings.IndexByte(part, '['); idx >= 0 {
				part = part[:idx]
			}
			inputs = append(inputs, ShaderInput{
				Name:      part,
				IsAttrib:  isAttrib,
				IsUniform: isUniform})
		}
	}
	return inputs
}

func shaderGetTopLevelLines(source string) []string {
	sourceBytes := []byte(source)
	l := len(sourceBytes)
	if l == 0 {
		return make([]string, 0)
	}

	var inPrecompiledStatement, inLineComment, inBlockComment, inString, stringEscape, lastWasWhitespace bool
	curlyBraceDepth := 0

	topLevelLines := make([]string, 0, 100)
	currentLine := make([]byte, 0, 1000)

	var c0, c1 byte = ' ', ' '
	for i := 0; i < l; i++ {
		c1 = sourceBytes[i]
		isWhitespace := unicode.IsSpace(rune(c1))
		if !inBlockComment && !inString && c0 == '/' && c1 == '/' {
			inLineComment = true
			if len(currentLine) > 0 {
				currentLine = currentLine[:len(currentLine)-1]
			}
		} else if !inBlockComment && !inString && c1 == '#' {
			inPrecompiledStatement = true
		} else if !inBlockComment && !inLineComment && !inPrecompiledStatement && !inString && c0 == '/' && c1 == '*' {
			inBlockComment = true
			if len(currentLine) > 0 {
				currentLine = currentLine[:len(currentLine)-1]
			}
		} else if !inBlockComment && !inString && !inLineComment && !inPrecompiledStatement && c1 == '"' {
			inString = true
		} else if inString && !stringEscape && c1 == '\\' {
			stringEscape = true
		} else if inString && stringEscape {
			stringEscape = false
		} else if inString && !stringEscape && c1 == '"' {
			inString = false
		} else if !inBlockComment && !inLineComment && !inPrecompiledStatement && !inString && c1 == '{' {
			if curlyBraceDepth == 0 {
				topLevelLines = append(topLevelLines, string(currentLine))
				currentLine = currentLine[:0]
			}
			curlyBraceDepth++
		}
		if !inBlockComment && !inLineComment && !inPrecompiledStatement && !inString && curlyBraceDepth == 0 {
			if c1 == ';' {
				topLevelLines = append(topLevelLines, string(currentLine))
				currentLine = currentLine[:0]
			} else if !isWhitespace {
				currentLine = append(currentLine, c1)
			} else if !lastWasWhitespace {
				currentLine = append(currentLine, ' ')
			}
		}
		if !inBlockComment && !inLineComment && !inPrecompiledStatement && !inString && c1 == '}' {
			curlyBraceDepth--
		} else if (inLineComment || inPrecompiledStatement) && (c1 == '\r' || c1 == '\n') {
			inLineComment = false
			inPrecompiledStatement = false
		} else if inBlockComment && c0 == '*' && c1 == '/' {
			inBlockComment = false
		}

		lastWasWhitespace = isWhitespace
		c0 = c1
	}

	topLevelLines = append(topLevelLines, string(currentLine))

	return topLevelLines
}

const compilePart = `
	{
		SHADER_VAR = gli.CreateShader(gl_SHADER_TYPE)
		gli.ShaderSource(SHADER_VAR, SHADER_SRC)
		gli.CompileShader(SHADER_VAR)

		shLog := gli.GetShaderInfoLog(SHADER_VAR)
		if len(shLog) > 0 {
			fmt.Printf("SHADER_TYPE compilation log for SHADER_SRC:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetShaderiv(SHADER_VAR, gl_COMPILE_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(SHADER_VAR)
			return nil, errors.New("Error compiling GL_SHADER_TYPE shader part for SHADER_SRC")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error compiling shader part for SHADER_SRC, glError: " + fmt.Sprint(glErr))
		}
	}
`
const linkPart = `
	{
		program = gli.CreateProgram()
		gli.AttachShader(program, vs)
		gli.AttachShader(program, fs)
		gli.LinkProgram(program)

		shLog := gli.GetProgramInfoLog(program)
		if len(shLog) > 0 {
			fmt.Printf("Shader link log for SHADER_SRC:\n\n%s\n", shLog)
		}

		var status int32
		gli.GetProgramiv(program, gl_LINK_STATUS, &status)
		if status != gl_TRUE {
			gli.DeleteShader(vs)
			gli.DeleteShader(fs)
			return nil, errors.New("error linking shader for SHADER_SRC")
		}
		if glErr := gli.GetError(); glErr != gl_NO_ERROR {
			return nil, errors.New("error linking shader for SHADER_SRC, glError: " + fmt.Sprint(glErr))
		}
	}
`

func capitalize(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func buildCodeHeader(buf *bytes.Buffer) {
	fmt.Fprint(buf, "package canvas\n\n")
	fmt.Fprint(buf, "import (\n")
	fmt.Fprint(buf, "\t\"errors\"\n")
	fmt.Fprint(buf, "\t\"fmt\"\n")
	fmt.Fprint(buf, ")\n\n")
}

func buildCode(buf *bytes.Buffer, baseName string, inputs []ShaderInput) {
	shaderName := baseName + "Shader"
	vsName := baseName + "VS"
	fsName := baseName + "FS"

	fmt.Fprintf(buf, "type %s struct {\n", shaderName)
	fmt.Fprint(buf, "\tid uint32\n")
	for _, input := range inputs {
		if input.IsAttrib {
			fmt.Fprintf(buf, "\t%s uint32\n", input.Name)
		} else if input.IsUniform {
			fmt.Fprintf(buf, "\t%s int32\n", input.Name)
		}
	}
	fmt.Fprint(buf, "}\n\n")
	fmt.Fprintf(buf, "func load%s() (*%s, error) {\n", capitalize(shaderName), shaderName)
	fmt.Fprint(buf, "\tvar vs, fs, program uint32\n")

	var part string

	part = strings.Replace(compilePart, "SHADER_SRC", vsName, -1)
	part = strings.Replace(part, "SHADER_TYPE", "VERTEX_SHADER", -1)
	part = strings.Replace(part, "SHADER_VAR", "vs", -1)
	fmt.Fprint(buf, part)

	part = strings.Replace(compilePart, "SHADER_SRC", fsName, -1)
	part = strings.Replace(part, "SHADER_TYPE", "FRAGMENT_SHADER", -1)
	part = strings.Replace(part, "SHADER_VAR", "fs", -1)
	fmt.Fprint(buf, part)

	part = strings.Replace(linkPart, "SHADER_SRC", fsName, -1)
	fmt.Fprint(buf, part)

	fmt.Fprint(buf, "\n")
	fmt.Fprintf(buf, "\tresult := &%s{}\n", shaderName)
	fmt.Fprint(buf, "\tresult.id = program\n")

	for _, input := range inputs {
		if input.IsAttrib {
			fmt.Fprintf(buf, "\tresult.%s = uint32(gli.GetAttribLocation(program, \"%s\"))\n", input.Name, input.Name)
		} else if input.IsUniform {
			fmt.Fprintf(buf, "\tresult.%s = gli.GetUniformLocation(program, \"%s\")\n", input.Name, input.Name)
		}
	}

	fmt.Fprint(buf, "\n")

	fmt.Fprint(buf, "\treturn result, nil\n")
	fmt.Fprint(buf, "}\n")
}

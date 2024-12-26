package render

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	core "github.com/gofiber/template"
	regex "github.com/tkdeng/goregex"
	"github.com/tkdeng/goutil"
	"github.com/tkdeng/staticweb"
)

type Engine struct {
	core.Engine

	src          string
	templatePath string
}

var varChar = [2]string{"{", "}"}

func SetVarChar(open, close string) {
	varChar[0] = open
	varChar[1] = close
}

func init() {
	staticweb.TemplateMode = true
}

// New returns a HTML render engine for Fiber
func New(directory string) (*Engine, error) {
	return newEngine(directory, nil)
}

// NewFileSystem returns a HTML render engine for Fiber with file system
func NewFileSystem(fs http.FileSystem) (*Engine, error) {
	return newEngine("/", fs)
}

// newEngine creates a new Engine instance with common initialization logic.
func newEngine(directory string, fs http.FileSystem) (*Engine, error) {
	out := filepath.Clean(directory) + ".dist"
	os.MkdirAll(out, 0755)

	err := staticweb.Compile(directory, out)
	if err != nil {
		return nil, err
	}

	engine := &Engine{
		Engine: core.Engine{
			Directory:  out,
			FileSystem: fs,
			Extension:  "html",
			Funcmap:    make(map[string]interface{}),
			Left:       varChar[0],
			Right:      varChar[1],
		},

		src:          directory,
		templatePath: out,
	}

	return engine, nil
}

// Load parses the templates to the engine.
func (e *Engine) Load() error {
	return staticweb.Compile(e.src, e.templatePath)
}

// Render will execute the template name along with the given values.
func (e *Engine) Render(out io.Writer, name string, binding interface{}, layout ...string) error {
	args := map[string]any{}
	if b, ok := binding.(map[string]any); ok {
		args = b
	} else if b, ok := binding.(fiber.Map); ok {
		args = b
	}

	var buf []byte
	var err error

	buf, err = getTemplateFile(e.templatePath, name, "index.html")
	if err != nil || len(buf) == 0 {
		buf, err = getTemplateFile(e.templatePath, name+".html")
	}

	if err != nil {
		return err
	}

	// compile args into vars
	buf = regex.Comp(`%1([\w_\-\.]+)%2`, e.Left, e.Right).RepFunc(buf, func(data func(int) []byte) []byte {
		varName := strings.Split(string(data(1)), ".")
		var varVal interface{}
		varVal = args

		for _, name := range varName {
			if val, ok := varVal.(map[string]any); ok {
				varVal = val[name]
			} else if val, ok := varVal.([]any); ok {
				if i, err := strconv.Atoi(name); err == nil {
					varVal = val[i]
				} else {
					break
				}
			} else {
				break
			}
		}

		return goutil.ToType[[]byte](varVal)
	})

	if _, err = out.Write(buf); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}

func getTemplateFile(name ...string) ([]byte, error) {
	path, err := goutil.JoinPath(name...)
	if err != nil {
		return []byte{}, err
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}

	return buf, nil
}

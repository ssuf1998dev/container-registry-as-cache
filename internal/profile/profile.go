package profile

import (
	"bytes"
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/goccy/go-yaml"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type Profile struct {
	Keys     []string `yaml:"keys"`
	DepFiles []string `yaml:"deps"`
	Files    []string `yaml:"files"`
}

//go:embed pnpm.yaml
var pnpm string
var ProfilePnpm Profile

func init() {
	rendered, _ := Render(pnpm)
	_ = yaml.Unmarshal(rendered, &ProfilePnpm)
}

var funcs = template.FuncMap{
	"sh": func(cmd string) (string, error) {
		file, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		runner, _ := interp.New(
			interp.StdIO(nil, &buf, &buf),
		)
		err = runner.Run(context.Background(), file)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	},
	"rel": filepath.Rel,
	"cwd": os.Getwd,
}

func Render(text string) ([]byte, error) {
	tpl, err := template.New("").Funcs(funcs).Funcs(sprig.FuncMap()).Parse(text)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

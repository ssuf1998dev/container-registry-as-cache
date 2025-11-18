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
	"github.com/goccy/go-yaml/ast"
	"github.com/ssuf1998dev/container-registry-as-cache/internal/utils"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type ProfileFiles struct {
	Value map[string]string
}

func (f *ProfileFiles) UnmarshalYAML(raw ast.Node) error {
	switch node := raw.(type) {
	case *ast.SequenceNode:
		patterns := []string{}
		iter := node.ArrayRange()
		for iter.Next() {
			if v, ok := iter.Value().(*ast.StringNode); ok {
				patterns = append(patterns, v.Value)
			}
		}

		f.Value = utils.ScanFiles(patterns)
	}
	return nil
}

type Profile struct {
	Keys     []string     `yaml:"keys"`
	DepFiles ProfileFiles `yaml:"deps"`
	Files    ProfileFiles `yaml:"files"`
}

//go:embed pnpm.yaml
var Pnpm string

func funcs(workdir string) template.FuncMap {
	return template.FuncMap{
		"sh": func(cmd string) (string, error) {
			file, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
			if err != nil {
				return "", err
			}
			var buf bytes.Buffer
			runner, _ := interp.New(
				interp.StdIO(nil, &buf, &buf),
				interp.Dir(workdir),
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
}

func Render(text string, workdir string) (*Profile, error) {
	tpl, err := template.New("").Funcs(funcs(workdir)).Funcs(sprig.FuncMap()).Parse(text)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, nil); err != nil {
		return nil, err
	}
	var p Profile
	if err := yaml.Unmarshal(buf.Bytes(), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

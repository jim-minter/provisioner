package config

import (
	"bytes"
	"os"
	"text/template"
)

func (config *Config) Template(tmpl string) ([]byte, error) {
	buf := &bytes.Buffer{}

	t, err := template.New("template").Funcs(map[string]any{
		"env": os.Getenv,
	}).Parse(tmpl)
	if err != nil {
		return nil, err
	}

	if err = t.Execute(buf, config); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

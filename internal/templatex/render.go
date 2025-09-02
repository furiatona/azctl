package templatex

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/furiatona/azctl/internal/config"
)

// RenderEnv replaces placeholders like {{VAR}} using values from Config.
func RenderEnv(input string, cfg *config.Config) (string, error) {
	// register functions before parsing
	t := template.New("aci").Option("missingkey=error").Funcs(template.FuncMap{
		"env": func(k string) (string, error) {
			v := cfg.Get(k)
			if v == "" {
				return "", fmt.Errorf("missing env: %s", k)
			}
			return v, nil
		},
	})
	var err error
	t, err = t.Parse(input)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

package template

import (
	"bytes"
	"fmt"
	"path"
	"text/template"

	"github.com/goccy/go-yaml"
	"karawale.in/go/lilac/microformats"
	"karawale.in/go/lilac/post"
)

func RenderMarkdown(entry post.Post) string {
	frontmatter, _ := yaml.Marshal(entry)

	md := fmt.Sprintf("---\n%s\n---\n%s", frontmatter, entry.Content.Html)
	return md
}

func RenderTemplate(entry post.Post, tmplpath string) (string, error) {
	result := bytes.NewBufferString("")
	funcMap := template.FuncMap(template.FuncMap{
		"yaml": func(m map[string]interface{}) (string, error) {
			marshaled, err := yaml.Marshal(m)
			if err != nil {
				return "", err
			}
			return string(marshaled), nil
		},
		"move": func(m map[string]interface{},
			from string, to string) map[string]interface{} {
			_, exists := m[from]
			if exists {
				m[to] = m[from]
				delete(m, from)
			}
			return m
		},
		"delete": func(m map[string]interface{},
			key string) map[string]interface{} {
			delete(m, key)
			return m
		},
	})

	tmpl, err := template.New("post").Funcs(funcMap).ParseFiles(tmplpath)
	if err != nil {
		return "", err
	}

	if err = tmpl.ExecuteTemplate(result, path.Base(tmplpath), struct {
		Post microformats.Jf2
	}{Post: post.PostToJf2(entry)}); err != nil {
		return "", err
	}
	return result.String(), nil
}

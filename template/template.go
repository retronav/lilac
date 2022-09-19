package template

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"karawale.in/go/lilac/post"
)

func RenderMarkdown(entry post.Post) string {
	frontmatter, _ := yaml.Marshal(entry)

	md := fmt.Sprintf("---\n%s\n---\n%s", frontmatter, entry.Content.Html)
	return md
}

package post

import (
	"github.com/fatih/structs"
	"github.com/k3a/html2text"
	"go.karawale.in/lilac/microformats"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Post struct {
	POST_TYPE string

	Type string `json:"type"`

	Content *struct {
		Html string `json:"html"`
		Text string `json:"text"`
	} `json:"content,omitempty"`
	Name       *string   `json:"name,omitempty"`
	Category   *[]string `json:"category,omitempty"`
	Published  *string   `json:"published,omitempty"`
	Updated    *string   `json:"updated,omitempty"`
	LikeOf     *string   `json:"like-of,omitempty"`
	RepostOf   *string   `json:"repost-of,omitempty"`
	BookmarkOf *string   `json:"bookmark-ok,omitempty"`
	InReplyTo  *string   `json:"in-reply-to,omitempty"`
}

func NormalizePostProperties(post microformats.Jf2) microformats.Jf2 {
	for key, value := range post {
		switch key {
		case "content":
			switch value := value.(type) {
			case string:
				post[key] = map[string]interface{}{
					"html": value,
					"text": value,
				}
			case map[string]interface{}:
				valueKeys := maps.Keys(value)
				if slices.Contains(valueKeys, "html") &&
					!slices.Contains(valueKeys, "text") {
					value["text"] = html2text.HTML2Text(value["html"].(string))
				} else if slices.Contains(valueKeys, "text") &&
					!slices.Contains(valueKeys, "html") {
					value["html"] = value["text"]
				}
			}
		}
	}

	return post
}

func PostToJf2(post Post) microformats.Jf2 {
	jf2 := make(microformats.Jf2)

	s := structs.New(post)
	s.TagName = "json"

	s.FillMap(jf2)
	return jf2
}

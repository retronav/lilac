package post

import (
	"reflect"

	"github.com/fatih/structs"
	"go.karawale.in/lilac/microformats"
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
		rValue := reflect.ValueOf(value)
		switch key {
		case "content":
			switch rValue.Type().Kind() {
			case reflect.String:
				post[key] = map[string]interface{}{
					"html": rValue.String(),
					"text": rValue.String(),
				}
			case reflect.Map:
				valueKeys := rValue.MapKeys()
				if slices.Contains(valueKeys, reflect.ValueOf("text")) &&
					!slices.Contains(valueKeys, reflect.ValueOf("html")) {
					post[key].(map[string]interface{})["html"] =
						rValue.MapIndex(reflect.ValueOf("text"))
				} else if slices.Contains(valueKeys, reflect.ValueOf("html")) &&
					!slices.Contains(valueKeys, reflect.ValueOf("text")) {
					post[key].(map[string]interface{})["text"] =
						rValue.MapIndex(reflect.ValueOf("html"))
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

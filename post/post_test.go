package post

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	"go.karawale.in/lilac/microformats"
)

func TestPost(t *testing.T) {
	jf2 := microformats.Jf2{
		"POST_type": "note",
		"type":      "entry",
		"content":   "hello world",
	}
	jf2 = NormalizePostProperties(jf2)
	post := Post{}
	mapstructure.Decode(jf2, &post)

	postjson, _ := json.Marshal(post)
	fmt.Println(string(postjson))
}

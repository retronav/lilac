package post

import (
	"fmt"
	"testing"

	"go.karawale.in/lilac/microformats"
)

func TestNormalizePostProperties(t *testing.T) {
	sample := microformats.Jf2{
		"content": map[string]interface{}{
			"text": "foo bar",
		},
	}
	fmt.Println(NormalizePostProperties(sample))
}

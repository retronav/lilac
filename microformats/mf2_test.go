package microformats

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/matryer/is"
)

func TestJf2ToMf2(t *testing.T) {
	is := is.New(t)
	fixtures := []fixture[Jf2, Mf2]{
		{
			Name: "checkin",
			Raw: Jf2{
				"type": "entry",
				"checkin": map[string]interface{}{
					"type":      "card",
					"name":      "Some place",
					"latitude":  "00",
					"longitude": "00",
				},
				"content":  "Looking nice",
				"category": []interface{}{"foo", "bar"},
			},
			Want: Mf2{
				Type: [1]string{"h-entry"},
				Properties: map[string]interface{}{
					"category": []interface{}{"foo", "bar"},
					"content":  []interface{}{"Looking nice"},
					"checkin": Mf2{
						Type: [1]string{"h-card"},
						Properties: map[string]interface{}{
							"latitude":  []interface{}{"00"},
							"longitude": []interface{}{"00"},
							"name":      []interface{}{"Some place"},
						},
					},
				},
			},
		},
		{
			Name: "Simple post",
			Raw: Jf2{
				"type":     "entry",
				"content":  "This is a simple post.",
				"category": []interface{}{"foo", "bar"},
			},
			Want: Mf2{
				Type: [1]string{"h-entry"},
				Properties: map[string]interface{}{
					"content":  []interface{}{"This is a simple post."},
					"category": []interface{}{"foo", "bar"},
				},
			},
		},
		{
			Name: "HTML Article",
			Raw: Jf2{
				"name": "Test article",
				"content": map[string]interface{}{
					"html": "<div>This is a <strong>test</strong> article fixture.</div>",
				},
				"category": []interface{}{"foo", "bar"},
				"mp-slug":  "test-article",
				"type":     "entry",
			},
			Want: Mf2{
				Type: [1]string{"h-entry"},
				Properties: map[string]interface{}{
					"content": map[string]interface{}{
						"html": []interface{}{"<div>This is a <strong>test</strong> article fixture.</div>"},
					},
					"category": []interface{}{"foo", "bar"},
					"mp-slug":  []interface{}{"test-article"},
					"name":     []interface{}{"Test article"},
				},
			},
		},
	}
	for _, f := range fixtures {
		t.Logf("fixture: %s", f.Name)
		got, err := Jf2ToMf2(f.Raw)
		is.NoErr(err)
		if diff := deep.Equal(got, f.Want); diff != nil {
			t.Error(diff)
		} else {
			t.Logf("success: %s", f.Name)
		}
	}
}

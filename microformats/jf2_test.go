package microformats

import (
	"net/url"
	"testing"

	"github.com/go-test/deep"
	"github.com/karlseguin/typed"
	"github.com/matryer/is"
	"golang.org/x/exp/maps"
)

type fixture[T any, U any] struct {
	Name string
	Raw  T
	Want U

	Add     Jf2
	Replace Jf2
	Delete  interface{}
}

func TestJsonToJf2(t *testing.T) {
	is := is.New(t)
	fixtures := []fixture[string, Jf2]{
		{
			Name: "Checkin",
			Raw: `{
      "type": ["h-entry"],
      "properties": {
        "checkin": [
          {
            "type": ["h-card"],
            "properties": {
              "name": ["Some place"],
              "latitude": ["00"],
              "longitude": ["00"]
            }
          }
        ],
        "content": ["Looking nice"],
        "category": ["foo", "bar"]
      }
    }
    `,
			Want: Jf2{
				"type": "entry",
				"checkin": Jf2{
					"type":      "card",
					"name":      "Some place",
					"latitude":  "00",
					"longitude": "00",
				},
				"content":  "Looking nice",
				"category": []interface{}{"foo", "bar"},
			},
		},
		{
			Name: "Simple post",
			Raw: `{
        "type": ["h-entry"],
        "properties": {
          "content": ["This is a simple post."],
          "category": ["foo", "bar"]
        }
      }
      `,
			Want: Jf2{
				"type":     "entry",
				"content":  "This is a simple post.",
				"category": []interface{}{"foo", "bar"},
			},
		},
		{
			Name: "HTML Article",
			Raw: `{
        "type": ["h-entry"],
        "properties": {
          "name": ["Test article"],
          "content": [
            {
              "html": "<div>This is a <strong>test</strong> article fixture.</div>"
            }
          ],
          "category": ["foo", "bar"],
          "mp-slug": ["test-article"]
        }
      }
      `,
			Want: Jf2{
				"name": "Test article",
				"content": Jf2{
					"html": "<div>This is a <strong>test</strong> article fixture.</div>",
				},
				"category": []interface{}{"foo", "bar"},
				"mp-slug":  "test-article",
				"type":     "entry",
			},
		},
	}

	for _, f := range fixtures {
		t.Logf("fixture: %s", f.Name)
		got, err := JsonToJf2(typed.Must([]byte(f.Raw)))
		is.NoErr(err)
		if diff := deep.Equal(got, f.Want); diff != nil {
			t.Error(diff)
		} else {
			t.Logf("success: %s", f.Name)
		}
	}
}

func TestFormEncodedToJf2(t *testing.T) {
	is := is.New(t)
	fixtures := []fixture[url.Values, Jf2]{
		{
			Name: "Simple post",
			Raw: url.Values{
				"h":       {"entry"},
				"content": {"Hello World"},
			},
			Want: Jf2{
				"type":    "entry",
				"content": "Hello World",
			},
		},
		{
			Name: "Post with categories",
			Raw: url.Values{
				"h":       {"entry"},
				"content": {"Hello World"},
				// will be "category[]" in real requests
				"category": {"foo", "bar"},
			},
			Want: Jf2{
				"type":     "entry",
				"content":  "Hello World",
				"category": []interface{}{"foo", "bar"},
			},
		},
	}

	for _, f := range fixtures {
		t.Logf("fixture: %s", f.Name)
		got, err := FormEncodedToJf2(f.Raw)
		is.NoErr(err)
		if diff := deep.Equal(got, f.Want); diff != nil {
			t.Error(diff)
		} else {
			t.Logf("success: %s", f.Name)
		}
	}
}

func TestAddInJf2(t *testing.T) {
	post := Jf2{
		"type":    "entry",
		"content": "Hello World",
	}

	postAfterAddition := Jf2{
		"type":     "entry",
		"content":  "Hello World",
		"category": []interface{}{"foo", "bar"},
	}

	post.Add(map[string]interface{}{
		"category": []interface{}{"foo", "bar"},
	})

	if diff := deep.Equal(post, postAfterAddition); diff != nil {
		t.Error(diff)
	}
}

func TestReplaceInJf2(t *testing.T) {
	post := Jf2{
		"type":     "entry",
		"content":  "Hello World",
		"category": []interface{}{"1", "2", "3"},
	}

	postAfterAddition := Jf2{
		"type":     "entry",
		"content":  "Hello World",
		"category": []interface{}{"foo", "bar"},
	}

	post.Replace(map[string]interface{}{
		"category": []interface{}{"foo", "bar"},
	})

	if diff := deep.Equal(post, postAfterAddition); diff != nil {
		t.Error(diff)
	}
}

func TestDeleteInJf2(t *testing.T) {
	fixtures := []fixture[Jf2, Jf2]{
		{
			Name: "Delete whole item",
			Raw: Jf2{
				"type":     "entry",
				"content":  "Hello World",
				"category": []interface{}{"foo", "bar"},
			},
			Want: Jf2{
				"type":    "entry",
				"content": "Hello World",
			},
			Delete: []string{"category"},
		},
		{
			Name: "Delete specific element from array",
			Raw: Jf2{
				"type":     "entry",
				"content":  "Hello World",
				"category": []interface{}{"foo", "bar"},
			},
			Want: Jf2{
				"type":     "entry",
				"content":  "Hello World",
				"category": []interface{}{"foo"},
			},
			Delete: map[string]interface{}{"category": []interface{}{"bar"}},
		},
	}
	for _, f := range fixtures {
		t.Logf("fixture: %s", f.Name)
		got := maps.Clone(f.Raw)
		got.Delete(f.Delete)
		if diff := deep.Equal(got, f.Want); diff != nil {
			t.Error(diff)
		} else {
			t.Logf("success: %s", f.Name)
		}
	}
}

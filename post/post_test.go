package post

import (
	"errors"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/matryer/is"
	"karawale.in/go/lilac/microformats"
)

type fixture[T any, U any] struct {
	Name      string
	Raw       T
	Want      U
	IsInvalid bool
}

func TestJf2ToPost(t *testing.T) {
	is := is.New(t)
	publishedTime := time.Now()
	fixtures := []fixture[microformats.Jf2, Post]{
		{
			Name: "Text note",
			Raw: microformats.Jf2{
				"type":      "entry",
				"content":   "Hello World",
				"published": publishedTime,
			},
			Want: Post{
				POST_TYPE: PostNote,
				Type:      "entry",
				Content: &PostContent{
					Html: "Hello World",
					Text: "Hello World",
				},
				Published: publishedTime,
			},
		},
		{
			Name: "HTML Note",
			Raw: microformats.Jf2{
				"type":      "entry",
				"content":   map[string]interface{}{"html": "<p>Hello World</p>"},
				"published": publishedTime,
			},
			Want: Post{
				POST_TYPE: PostNote,
				Type:      "entry",
				Content: &PostContent{
					Html: "<p>Hello World</p>",
					Text: "Hello World",
				},
				Published: publishedTime,
			},
		},
		{
			Name: "HTML Article",
			Raw: microformats.Jf2{
				"type": "entry",
				"name": "Hello World",
				"content": map[string]interface{}{
					"html": "<p>Hello World. This is a <em>nice</em> article.</p>"},
				"published": publishedTime,
			},
			Want: Post{
				POST_TYPE: PostArticle,
				Type:      "entry",
				Content: &PostContent{
					Html: "<p>Hello World. This is a <em>nice</em> article.</p>",
					Text: "Hello World. This is a nice article.",
				},
				Name:      "Hello World",
				Published: publishedTime,
			},
		},
		{
			Name: "Text note with location",
			Raw: microformats.Jf2{
				"type":      "entry",
				"content":   "Mr Worldwide",
				"published": publishedTime,
				"location":  "geo:00,00",
			},
			Want: Post{
				POST_TYPE: PostNote,
				Type:      "entry",
				Content: &PostContent{
					Html: "Mr Worldwide",
					Text: "Mr Worldwide",
				},
				Location: &PostLocation{
					Latitude:  "00",
					Longitude: "00",
				},
				Published: publishedTime,
			},
		},
		{
			Name: "Checkin",
			Raw: microformats.Jf2{
				"type":      "entry",
				"content":   "Checked in at this place",
				"published": publishedTime,
				"checkin": map[string]interface{}{
					"type":      "card",
					"name":      "Some place",
					"latitude":  "00",
					"longitude": "00",
				},
				"category": "travel",
			},
			Want: Post{
				POST_TYPE: PostCheckin,
				Type:      "entry",
				Content: &PostContent{
					Html: "Checked in at this place",
					Text: "Checked in at this place",
				},
				Checkin: &Checkin{
					Type:      "card",
					Name:      "Some place",
					Latitude:  "00",
					Longitude: "00",
				},
				Category:  []string{"travel"},
				Published: publishedTime,
			},
		},
	}

	for _, f := range fixtures {
		t.Logf("fixture: %s", f.Name)
		got, err := Jf2ToPost(f.Raw)
		if !f.IsInvalid {
			is.NoErr(err)
		} else {
			if !errors.Is(err, ErrorInvalidPost) {
				t.Error("Wanted an invalid post but got error: %w", err)
			}
		}
		if diff := deep.Equal(got, f.Want); diff != nil {
			t.Error(diff)
		} else {
			t.Logf("success: %s", f.Name)
		}
	}
}

func TestPostToJf2(t *testing.T) {
	publishedTime := time.Now()
	// TODO: add more tests
	fixtures := []fixture[Post, microformats.Jf2]{
		{
			Name: "Checkin",
			Raw: Post{
				POST_TYPE: PostCheckin,
				Type:      "entry",
				Content: &PostContent{
					Html: "Checked in at this place",
					Text: "Checked in at this place",
				},
				Checkin: &Checkin{
					Type:      "card",
					Name:      "Some place",
					Latitude:  "00",
					Longitude: "00",
				},
				Category:  []string{"foo", "bar"},
				Published: publishedTime,
			},
			Want: microformats.Jf2{
				"POST_TYPE": PostCheckin,
				"type":      "entry",
				"content": map[string]interface{}{
					"html": "Checked in at this place",
					"text": "Checked in at this place",
				},
				"published": publishedTime,
				"checkin": map[string]interface{}{
					"type":      "card",
					"name":      "Some place",
					"latitude":  "00",
					"longitude": "00",
				},
				"category": []string{"foo", "bar"},
			},
		},
	}

	for _, f := range fixtures {
		t.Logf("fixture: %s", f.Name)
		got := PostToJf2(f.Raw)
		if diff := deep.Equal(got, f.Want); diff != nil {
			t.Error(diff)
		} else {
			t.Logf("success: %s", f.Name)
		}
	}
}

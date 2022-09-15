package post

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/go-playground/validator/v10"
	"github.com/k3a/html2text"
	"github.com/mitchellh/mapstructure"
	"go.karawale.in/lilac/microformats"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	// ErrorInvalidPost is used when a post (usually during normalizing in its
	// Jf2 form) has bad values. An annotation must be added to the error before
	// returning, explaining what makes the post invalid.
	ErrorInvalidPost = errors.New("invalid post")
)

// PostType is a type alias used for post types.
type PostType string

// Various post types which go in Post.POST_TYPE.
const (
	PostNote     PostType = "note"
	PostArticle  PostType = "article"
	PostLike     PostType = "like"
	PostReply    PostType = "reply"
	PostRepost   PostType = "repost"
	PostBookmark PostType = "bookmark"
	PostCheckin  PostType = "checkin"
)

// GeoUriRe is a regular expression to extract latitude, longitude and altitude
// from Geo URIs (RFC 5870).
var GeoUriRe = regexp.MustCompile(`geo:(?P<lat>[-?\d+.]*),(?P<lon>[-?\d+.]*)(,(?P<alt>[-?\d+.]*))?`)

// Post is a structured format of a post which makes modifying/accessing
// properties easier than the raw mf2/jf2 form.
type Post struct {
	POST_TYPE PostType `validate:"required"`

	Type      string    `json:"type" validate:"required"`
	Published time.Time `json:"published,omitempty" validate:"required"`

	Content    *PostContent  `json:"content,omitempty"`
	Location   *PostLocation `json:"location,omitempty"`
	Photo      *[]PostPhoto  `json:"photo,omitempty"`
	Checkin    *Checkin      `json:"checkin,omitempty"`
	Name       string        `json:"name,omitempty"`
	Category   []string      `json:"category,omitempty"`
	Updated    time.Time     `json:"updated,omitempty"`
	LikeOf     string        `json:"like-of,omitempty"`
	RepostOf   string        `json:"repost-of,omitempty"`
	BookmarkOf string        `json:"bookmark-ok,omitempty"`
	InReplyTo  string        `json:"in-reply-to,omitempty"`
	Slug       string        `json:"mp-slug,omitempty"`
}

// normalizeJf2Post prepares the jf2 form of a post to be converted to the
// struct Post.
func normalizeJf2Post(post microformats.Jf2) (microformats.Jf2, error) {
	postKeys := maps.Keys(post)
	for _, key := range postKeys {
		value := post[key]
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
			default:
				return nil, ErrorInvalidPost
			}
		case "location":
			switch value := value.(type) {
			case string:
				if strings.HasPrefix(value, "geo:") && GeoUriRe.MatchString(value) {
					// Location is a Geo URI
					matches := GeoUriRe.FindStringSubmatch(value)
					latitude := matches[GeoUriRe.SubexpIndex("lat")]
					longitude := matches[GeoUriRe.SubexpIndex("lon")]
					altitude := matches[GeoUriRe.SubexpIndex("alt")]
					if latitude != "" && longitude != "" {
						location := map[string]interface{}{
							"latitude":  latitude,
							"longitude": longitude,
						}
						if altitude != "" {
							location["altitude"] = altitude
						}
						post[key] = location
					}
				} else {
					// No idea what this is
					return nil, fmt.Errorf("%w: invalid location string value", ErrorInvalidPost)
				}
			default:
				return nil, ErrorInvalidPost
			}
		case "photo":
			if value, ok := value.(string); ok {
				post[key] = []map[string]string{
					{"value": value, "alt": ""},
				}
			}
			if value, ok := value.([]string); ok {
				photos := []map[string]string{}
				for _, photo := range value {
					photos = append(photos, map[string]string{"value": photo, "alt": ""})
				}
				post[key] = photos
			}
		case "slug":
			// slug is deprecated, move it to mp-slug
			delete(post, "slug")
			post["mp-slug"] = value.(string)
		}
	}
	if !slices.Contains(postKeys, "published") {
		post["published"] = time.Now()
	}
	return post, nil
}

func getPostType(post microformats.Jf2) PostType {
	keys := maps.Keys(post)

	keysToType := map[string]PostType{
		"in-reply-to": PostReply,
		"bookmark-of": PostBookmark,
		"like-of":     PostLike,
		"repost-of":   PostRepost,
		"name":        PostArticle,
		"checkin":     PostCheckin,
	}

	for key, postType := range keysToType {
		if slices.Contains(keys, key) {
			return postType
		}
	}

	return PostNote
}

// PostToJf2 marshals a post (struct) back to its jf2 form.
func PostToJf2(post Post) microformats.Jf2 {
	jf2 := make(microformats.Jf2)

	s := structs.New(post)
	s.TagName = "json"

	s.FillMap(jf2)

	// de-reference any top level pointers
	for key, value := range jf2 {
		rvalue := reflect.ValueOf(value)
		if rvalue.Kind() == reflect.Pointer && !rvalue.IsNil() {
			jf2[key] = *value.(*interface{})
		}
	}

	return jf2
}

// Jf2ToPost unmarshals a jf2 post to the struct Post.
func Jf2ToPost(jf2 microformats.Jf2) (Post, error) {
	jf2 = maps.Clone(jf2)

	var post Post
	normalized, err := normalizeJf2Post(jf2)
	if err != nil {
		return post, err
	}
	normalized["POST_TYPE"] = getPostType(normalized)
	if err = mapstructure.Decode(normalized, &post); err != nil {
		return post, fmt.Errorf("%w: %s", ErrorInvalidPost, err.Error())
	}
	validate := validator.New()
	if err = validate.Struct(post); err != nil {
		return post, fmt.Errorf("%w: %s", ErrorInvalidPost, err.Error())
	}
	return post, nil
}

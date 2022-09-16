package post

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/fatih/structs"
	"github.com/go-playground/validator/v10"
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

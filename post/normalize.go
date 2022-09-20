package post

import (
	"fmt"
	"strings"
	"time"

	"github.com/k3a/html2text"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"karawale.in/go/lilac/microformats"
)

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
			switch value := value.(type) {
			case string:
				post[key] = []map[string]string{
					{"value": value, "alt": ""},
				}
			case []string:
				photos := []map[string]string{}
				for _, photo := range value {
					photos = append(photos, map[string]string{"value": photo, "alt": ""})
				}
				post[key] = photos
			case map[string]string:
				// a single photo will be flattened; reverse it
				post[key] = []map[string]string{value}

			}
		case "category":
			switch value := value.(type) {
			// a single category will be flattened; reverse it
			case string:
				post[key] = []string{value}
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

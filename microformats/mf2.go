package microformats

import (
	"reflect"

	"golang.org/x/exp/slices"
)

// Mf2 is a microformats2 entry.
//
// Nested properties in Jf2 should be loosely typed as map[string]interface{} or
// []interface{} wherever applicable.
type Mf2 struct {
	Type       [1]string              `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// Reserved properties in microformats2 object. DO NOT MUTATE.
var RESERVED_PROPERTIES = []string{
	"h",
	"access_token",
	"action",
	"url",
}

// Reserved properties in microformats2 object for internal use in Lilac.
// DO NOT MUTATE.
var RESERVED_PROPERTIES_FOR_LILAC = []string{
	"POST_TYPE",
}

// Jf2ToMf2 converts a jf2 representation of a entry back to microformats2.
func Jf2ToMf2(jf2 Jf2) (Mf2, error) {
	mf2 := Mf2{
		Type:       [1]string{"h-" + jf2["type"].(string)},
		Properties: make(map[string]interface{}),
	}
	err := nestProperties(&mf2, jf2)
	if err != nil {
		return mf2, err
	}
	return mf2, nil
}

func nestProperties(mf2 *Mf2, jf2 Jf2) error {
	for key, value := range jf2 {
		if slices.Contains(append(
			RESERVED_PROPERTIES, append(
				RESERVED_PROPERTIES_FOR_LILAC, "type")...), key) {
			continue
		}
		valueType := reflect.TypeOf(value)
		switch valueType.Kind() {
		case reflect.Map:
			value := value.(map[string]interface{})
			if value["type"] != nil {
				childMf2, err := Jf2ToMf2(value)
				if err != nil {
					return err
				}
				mf2.Properties[key] = childMf2
			} else {
				valueMap := make(map[string]interface{})
				for k, v := range value {
					if _, ok := v.([]interface{}); !ok {
						valueMap[k] = []interface{}{v}
					} else {
						valueMap[k] = v
					}
				}
				mf2.Properties[key] = valueMap
			}
		case reflect.Array, reflect.Slice:
			mf2.Properties[key] = value
		default:
			mf2.Properties[key] = []interface{}{value}
		}
	}
	return nil
}

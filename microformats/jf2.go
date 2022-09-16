package microformats

import (
	"errors"
	"reflect"
	"strings"

	"github.com/karlseguin/typed"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Jf2 is a jf2 representation of a microformats2 entry. See:
// https://microformats.org/wiki/jf2
//
// Nested properties in Jf2 should be loosely typed as map[string]interface{} or
// []interface{} wherever applicable.
type Jf2 map[string]interface{}

// Add adds properties to an existing Jf2 object.
// See: https://www.w3.org/TR/micropub/#h-add
func (j Jf2) Add(spec map[string]interface{}) {
	keys := maps.Keys(j)
	for key, value := range spec {
		if value, ok := value.([]interface{}); ok {
			if slices.Contains(keys, key) {
				j[key] = append(value, j[key])
			} else {
				if len(value) == 1 {
					j[key] = value[0]
				} else {
					j[key] = value
				}
			}
		}
	}
}

// Replace replaces properties in an existing Jf2 object.
// See: https://www.w3.org/TR/micropub/#h-replace
func (j Jf2) Replace(spec map[string]interface{}) {
	flatSpec := maps.Clone(spec)
	flattenValues(flatSpec)
	keys := maps.Keys(j)
	for key, value := range flatSpec {
		if slices.Contains(keys, key) {
			j[key] = value
		}
	}
}

// Delete deletes properties from an existing Jf2 object.
// See: https://www.w3.org/TR/micropub/#h-remove
func (j Jf2) Delete(spec interface{}) {
	switch spec := spec.(type) {
	case []string:
		maps.DeleteFunc(j, func(key string, value interface{}) bool {
			return slices.Contains(spec, key)
		})
	case map[string]interface{}:
		keys := maps.Keys(j)
		for key, value := range spec {
			if !slices.Contains(keys, key) {
				continue
			}
			value, ok := value.([]interface{})
			if !ok {
				continue
			}
			jvalue, ok := j[key].([]interface{})
			if !ok {
				continue
			}
			for i, el := range jvalue {
				for _, toDelete := range value {
					if el == toDelete {
						jvalue = slices.Delete(jvalue, i, i+1)
					}
				}
			}
			j[key] = jvalue
		}
	}
}

// JsonToJf2 converts JSON microformat entry to jf2.
func JsonToJf2(typed typed.Typed) (Jf2, error) {
	jf2 := Jf2{}

	entryType, ok := typed.StringsIf("type")
	if !ok || (entryType != nil && len(entryType) == 0) {
		return nil, errors.New("entry does not have a vaild type")
	}
	jf2["type"] = strings.TrimLeft(entryType[0], "h-")
	props, err := flattenProperties(jf2, typed)
	if err != nil {
		return nil, err
	}
	for key, value := range props {
		jf2[key] = value
	}
	return jf2, nil
}

// FormEncodedToJf2 converts the parsed map of form content to jf2.
func FormEncodedToJf2(form map[string][]string) (Jf2, error) {
	jf2 := make(Jf2)
	for key, value := range form {
		if key == "h" {
			key = "type"
		} else if slices.Contains(RESERVED_PROPERTIES, key) {
			continue
		}
		if len(value) == 1 {
			jf2[key] = value[0]
		} else {
			// Value is []string but is converted to []interface{} to align
			// behaviour with JsonToJf2.notes
			sliceValue := make([]interface{}, 0)
			for _, el := range value {
				sliceValue = append(sliceValue, el)
			}
			jf2[key] = sliceValue
		}
	}
	return jf2, nil
}

// flattenProperties recursively flattens the value of the "properties" key of a
// microformat entry.
func flattenProperties(jf2 Jf2, mf2 typed.Typed) (Jf2, error) {
	flattened := Jf2{}

	properties, ok := mf2.ObjectIf("properties")
	if !ok {
		return nil, errors.New("invalid properties field in mf2")
	}

	for _, key := range properties.Keys() {
		if slices.Contains(RESERVED_PROPERTIES, key) {
			continue
		}

		value := properties[key].([]interface{})
		if len(value) == 0 {
			continue
		}

		valueType := reflect.TypeOf(value[0])
		switch valueType.Kind() {
		case reflect.Map:
			// HACK: This works with only the first map in the slice and ignores
			// the rest. Need to have a look at this.
			valueMap := value[0].(map[string]interface{})
			if _, ok := valueMap["properties"]; ok {
				flattenedValue, err := JsonToJf2(typed.New(valueMap))
				if err != nil {
					return nil, err
				}
				flattened[key] = flattenedValue
			} else {
				flatValueMap := make(Jf2)
				for key, value := range valueMap {
					flatValueMap[key] = value
				}
				flattenValues(flatValueMap)
				flattened[key] = flatValueMap
			}
		default:
			if len(value) == 1 {
				flattened[key] = value[0]
			} else {
				flattened[key] = value
			}
		}
	}

	return flattened, nil
}

// flattenValues flattens single-element arrays to the element itself.
//
// Example: `["foo"]` => `"foo"`
func flattenValues(m map[string]interface{}) {
	for key, value := range m {
		if vSlice, ok := value.([]interface{}); ok {
			if len(vSlice) == 1 {
				m[key] = vSlice[0]
			}
		} else if child, ok := value.(map[string]interface{}); ok {
			child = maps.Clone(child)
			flattenValues(child)
			m[key] = child
		}
	}
}

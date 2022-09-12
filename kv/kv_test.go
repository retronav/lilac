package kv

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/matryer/is"
)

func TestKeyValueLoad(t *testing.T) {
	is := is.New(t)

	store := KeyValue[map[string]string]{}

	rawStore := `{"foo": "bar"}`
	expectedStoreContent := map[string]string{
		"foo": "bar",
	}

	err := store.Load([]byte(rawStore))
	is.NoErr(err)
	is.Equal(store.Content, expectedStoreContent)
}

func TestKeyValueSync(t *testing.T) {
	is := is.New(t)

	store := KeyValue[map[string]string]{
		Content: map[string]string{
			"foo": "bar",
		},
	}
	rawStore := new(bytes.Buffer)

	err := store.Sync(rawStore)
	is.NoErr(err)

	var decodedRawStore map[string]string
	err = json.Unmarshal(rawStore.Bytes(), &decodedRawStore)
	is.NoErr(err)
	is.Equal(decodedRawStore, store.Content)
}

func TestKeyValueMutability(t *testing.T) {
	is := is.New(t)

	store := KeyValue[map[string]string]{
		Content: map[string]string{
			"foo": "bar",
		},
	}
	mutateStore(&store)
	rawStore := new(bytes.Buffer)

	err := store.Sync(rawStore)
	is.NoErr(err)

	expectedDecodedStore := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}
	var decodedRawStore map[string]string
	err = json.Unmarshal(rawStore.Bytes(), &decodedRawStore)
	is.NoErr(err)
	is.Equal(decodedRawStore, expectedDecodedStore)
}

func mutateStore(store *KeyValue[map[string]string]) {
	store.Content["bar"] = "baz"
}

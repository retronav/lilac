package kv

import (
	"encoding/json"
	"io"
)

// KeyValue is a generic abstraction of a JSON key-value store.
type KeyValue[T any] struct {
	Content T
}

// Load loads raw JSON into the store.
func (k *KeyValue[T]) Load(store []byte) error {
	return json.Unmarshal(store, &k.Content)
}

// Sync marshals the store into JSON and writes it.
func (k *KeyValue[T]) Sync(writer io.Writer) error {
	store, err := json.Marshal(k.Content)
	if err != nil {
		return err
	}
	_, err = writer.Write(store)
	if err != nil {
		return err
	}
	return nil
}

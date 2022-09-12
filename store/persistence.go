package store

import (
	"bytes"
	"io"
	"path"

	log "github.com/sirupsen/logrus"
	"go.karawale.in/lilac/kv"
	"go.karawale.in/lilac/post"
)

// Persistence is a umbrella store of multiple KeyValue stores storing various
// things which can be needed for server operations.
type Persistence struct {
	// Path is the directory where persistence is located. (relative to the
	// content store).
	Path string
	// Store is the main content store used by persistence.
	Store *GitStore

	postMappingsFile   string
	postPropertiesFile string

	// PostMappings is a map of post web urls to location of markdown post file
	// on the store.
	PostMappings kv.KeyValue[map[string]string]
	// PostProperties is a map of post web urls to post properties.
	PostProperties kv.KeyValue[map[string]post.Post]
}

// NewPersistence returns a new instance of Persistence.
func NewPersistence(store *GitStore, location *string) Persistence {
	if location == nil {
		defaultLoc := ".lilac"
		location = &defaultLoc
	}

	p := Persistence{
		Path:  *location,
		Store: store,

		postMappingsFile:   path.Join(*location, "post_mappings.json"),
		postPropertiesFile: path.Join(*location, "post_properties.json"),

		PostMappings:   kv.KeyValue[map[string]string]{},
		PostProperties: kv.KeyValue[map[string]post.Post]{},
	}
	return p
}

// Load loads data from the content store into the instance.
func (p *Persistence) Load() error {
	file, err := p.Store.Read(p.postMappingsFile)
	if err != nil {
		return err
	}
	store, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = p.PostMappings.Load(store)
	if err != nil {
		return err
	}

	file, err = p.Store.Read(p.postPropertiesFile)
	if err != nil {
		return err
	}
	store, err = io.ReadAll(file)
	if err != nil {
		return err
	}
	err = p.PostProperties.Load(store)
	if err != nil {
		return err
	}

	log.Info("persistence initalized successfully")
	return nil
}

// Dump dumps data to the content store.
func (p Persistence) Dump() error {
	rawStore := new(bytes.Buffer)
	err := p.PostMappings.Sync(rawStore)
	if err != nil {
		return err
	}
	err = p.Store.Write(p.postMappingsFile, rawStore)
	if err != nil {
		return err
	}

	rawStore = new(bytes.Buffer)
	err = p.PostProperties.Sync(rawStore)
	if err != nil {
		return err
	}
	err = p.Store.Write(p.postPropertiesFile, rawStore)
	if err != nil {
		return err
	}

	return nil
}

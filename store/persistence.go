package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
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
func NewPersistence(store *GitStore, location string) Persistence {
	if location == "" {
		location = ".lilac"
	}

	if _, err := os.Stat(path.Join(store.Path, location)); errors.Is(err, os.ErrNotExist) {
		if err = store.Fs.Mkdir(location, 0777); err != nil {
			panic(fmt.Errorf("failed to create persistence dir: %w", err))
		}
	}

	p := Persistence{
		Path:  location,
		Store: store,

		postMappingsFile:   path.Join(location, "post_mappings.json"),
		postPropertiesFile: path.Join(location, "post_properties.json"),

		PostMappings:   kv.KeyValue[map[string]string]{},
		PostProperties: kv.KeyValue[map[string]post.Post]{},
	}
	return p
}

// Load loads data from the content store into the instance.
func (p *Persistence) Load() error {
	for _, filePath := range []string{p.postMappingsFile, p.postPropertiesFile} {
		if _, err := p.Store.Fs.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			file, err := p.Store.Fs.Create(filePath)
			if err != nil {
				return err
			}
			if _, err = file.WriteString("{}"); err != nil {
				return err
			}
			if err = file.Close(); err != nil {
				return err
			}
		}
	}

	file, err := p.Store.Fs.OpenFile(p.postMappingsFile, os.O_RDONLY, 0666)
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

	file, err = p.Store.Fs.OpenFile(p.postPropertiesFile, os.O_RDONLY, 0666)
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
	writeMap := map[string]func(io.Writer) error{
		p.postMappingsFile:   p.PostMappings.Sync,
		p.postPropertiesFile: p.PostProperties.Sync,
	}

	for filepath, syncer := range writeMap {
		rawStore := new(bytes.Buffer)
		err := syncer(rawStore)
		if err != nil {
			return err
		}
		file, err := p.Store.Fs.OpenFile(filepath, os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		if err := file.Truncate(0); err != nil {
			return err
		}
		_, err = file.Write(rawStore.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

package store

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"
)

// GitStore is a store implementation based on a local git repository. The store
// syncs changes to the remote at once i.e indivivual file operations are not
// directly reflected in the remote and must be synced using `GitStore.Sync`.
//
// All file operation functions accept file paths relative to the location of
// the git repository.
type GitStore struct {
	// Path is the relative path to the repository.
	Path string
	// Branch is a branch name in the repository which is to be used for
	// initializing the store.
	Branch string

	// wt is the initialized worktree from Init.
	wt *git.Worktree
}

// Init initializes the store. Init MUST be called before performing any
// operations.
func (g *GitStore) Init() error {
	absPath, err := filepath.Abs(g.Path)
	if err != nil {
		return err
	}
	g.Path = absPath

	log.Debugf("initializing git store at %s", g.Path)
	repo, err := git.PlainOpen(g.Path)
	if err != nil {
		return fmt.Errorf("repository error: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree error: %w", err)
	}

	_, err = repo.Branch(g.Branch)
	if err != nil {
		return fmt.Errorf("branch error: %w", err)
	}
	log.Debugf("branch %s exists", g.Branch)

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(g.Branch),
	})
	if err != nil {
		return fmt.Errorf("checkout error: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("cannot fetch repo status: %w", err)
	}

	if !status.IsClean() {
		err = wt.Pull(&git.PullOptions{})
		if err != nil {
			return fmt.Errorf("git force pull error: %w", err)
		}
		log.Debug("pulled updates")
	}
	g.wt = wt
	log.Info("git store initialized successfully")
	return nil
}

// Write writes to a file with content. If the file does not exist, it is
// created.
func (g GitStore) Write(filepath string, content io.Reader) error {
	filepath = path.Join(g.Path, filepath)
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, content)
	if err != nil {
		return err
	}
	log.Debugf("created file %s", filepath)
	return nil
}

// Delete deletes an existing file.
func (g GitStore) Delete(filepath string) error {
	filepath = path.Join(g.Path, filepath)
	err := os.Remove(filepath)
	if err != nil {
		return err
	}
	log.Debugf("deleted file %s", filepath)
	return nil
}

// Read returns the content of an existing file.
func (g GitStore) Read(filepath string) (io.Reader, error) {
	filepath = path.Join(g.Path, filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	log.Debugf("fetched file %s", filepath)
	return file, nil
}

// List recursively lists all the files in a directory.
func (g GitStore) List(dirpath string) ([]string, error) {
	dirpath = path.Join(g.Path, dirpath)
	paths := make([]string, 0)
	err := filepath.Walk(
		dirpath,
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				paths = append(paths, path)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

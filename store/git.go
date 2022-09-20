package store

import (
	"fmt"
	"path/filepath"
	"time"

	git "github.com/gogs/git-module"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
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

	// repo is the git repository.
	repo *git.Repository

	// Fs is the initialized filesystem.
	Fs afero.Fs
}

// Init initializes the store. Init MUST be called before performing any
// operations.
func (g *GitStore) Init() error {
	absPath, err := filepath.Abs(g.Path)
	if err != nil {
		return err
	}
	g.Path = absPath
	repo, err := git.Open(g.Path)
	if err != nil {
		return fmt.Errorf("error opening repository: %w", err)
	}
	if !repo.HasBranch(g.Branch) {
		return fmt.Errorf("branch %s does not exist", g.Branch)
	}
	if err = repo.Pull(git.PullOptions{
		All:    false,
		Remote: "origin",
		Branch: g.Branch}); err != nil {
		return err
	}
	log.Debugf("branch %s exists", g.Branch)
	log.Debugf("initializing git store at %s", g.Path)
	g.repo = repo
	g.Fs = afero.NewBasePathFs(afero.NewOsFs(), absPath)

	log.Infof("git store initialized successfully at %s", absPath)
	return nil
}

func (g GitStore) Sync(message string) error {
	if err := g.repo.Add(git.AddOptions{All: true}); err != nil {
		return err
	}
	committer := &git.Signature{
		Name:  "Lilac",
		Email: "lilac@localhost",
		When:  time.Now().UTC(),
	}
	if err := g.repo.Commit(committer, message); err != nil {
		return err
	}
	if err := g.repo.Push("origin", g.Branch); err != nil {
		return err
	}

	return nil
}

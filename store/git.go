package store

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
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

	// wt is the initialized worktree from Init.
	wt *git.Worktree

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
	g.repo = repo

	g.Fs = afero.NewBasePathFs(afero.NewOsFs(), absPath)

	log.Infof("git store initialized successfully at %s", absPath)
	return nil
}

func (g GitStore) Sync(message string) error {
	if err := g.wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return err
	}
	if _, err := g.wt.Commit(message, &git.CommitOptions{}); err != nil {
		return err
	}
	// HACK: go-git does not seem to be running hooks, use the cli instead
	cmd := exec.Command("git", "push")
	cmd.Dir = g.Path

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	in := bufio.NewScanner(io.MultiReader(stdout, stderr))
	logrus.Info("git push log:\n")
	for in.Scan() {
		logrus.Info(in.Text())
	}
	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

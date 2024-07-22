package resources

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"path/filepath"
	"strings"
	"time"
)

type GitModule struct {
	Name       string `hcl:"name,label"`
	Repository string `hcl:"repository"`
	Ref        string `hcl:"ref"`
	WorkingDir string `hcl:"working_dir,optional"`
}

func (g *GitModule) Perform(e *ModuleEnv) error {
	ctx, done := context.WithTimeout(context.Background(), 30*time.Second)
	defer done()

	var err error
	var repo *git.Repository
	if strings.HasPrefix(g.Repository, "/") || strings.HasPrefix(g.Repository, "file://") {
		repoRoot, err := e.source.Chroot(g.Repository)
		if err != nil {
			return err
		}
		atticRoot, err := repoRoot.Chroot(".git")
		if err != nil {
			return err
		}

		wc := memfs.New()
		gitAttic := filesystem.NewStorage(atticRoot, cache.NewObjectLRUDefault())
		repo, err = git.Open(gitAttic, wc)
		if err != nil {
			return err
		}
		ref, _ := repo.Reference(plumbing.NewBranchReferenceName(g.Ref), true)
		w, _ := repo.Worktree()
		if err := w.Reset(&git.ResetOptions{
			Commit: ref.Hash(),
			Mode:   git.HardReset,
		}); err != nil {
			return err
		}
		dumpFileSystem(wc, "/")
	} else {
		gitAttic := memory.NewStorage()
		workingTree := memfs.New()
		repo, err = git.CloneContext(ctx, gitAttic, workingTree, &git.CloneOptions{
			URL:           g.Repository,
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(g.Ref),
		})
		if err != nil {
			fmt.Printf("Clone failed %e\n", err)
			return err
		}
	}
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	configRoot := "/"
	if g.WorkingDir != "" {
		configRoot = g.WorkingDir
	}

	if err := e.ApplySubmodule(tree.Filesystem, configRoot); err != nil {
		return err
	}

	e.Completed("git_module", g.Name)
	return nil
}

func dumpFileSystem(fs billy.Filesystem, path string) {
	info, err := fs.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, file := range info {
		filePath := filepath.Join(path, file.Name())
		fmt.Printf("%s\n", filePath)
		if file.IsDir() {
			dumpFileSystem(fs, filePath)
		}
	}
}

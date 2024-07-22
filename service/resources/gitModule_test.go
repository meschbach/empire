package resources

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGitModule(t *testing.T) {
	t.Run("Runs a module from a local git repository", func(t *testing.T) {
		t.Skip()
		sourceFS := memfs.New()
		targetFS := memfs.New()
		//setup the repository
		repoFS, err := sourceFS.Chroot("/git-example/.git")
		require.NoError(t, err)
		repo, err := git.InitWithOptions(filesystem.NewStorage(repoFS, cache.NewObjectLRUDefault()), repoFS, git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName("main"),
		})
		require.NoError(t, err)
		tree, err := repo.Worktree()
		require.NoError(t, err)
		writeTestFile(t, tree.Filesystem, "base.hcl", "static_local_file {\nsource = \"input\"\ntarget = \"/var/lib/output\"\n}")
		_, err = tree.Add("base.hcl")
		require.NoError(t, err)
		writeTestFile(t, tree.Filesystem, "input", "flag")
		_, err = tree.Add("input")
		require.NoError(t, err)
		_, err = tree.Commit("Init", &git.CommitOptions{})
		require.NoError(t, err)

		cfg := File{
			GitModule: []GitModule{
				{
					Name:       "subject-under-test",
					Repository: "/git-example",
					Ref:        "main",
				},
			},
		}
		env := NewModuleEnv(sourceFS, "/etc/empire", targetFS, "/", nil)
		require.NoError(t, env.resolve(&cfg))

		fmt.Printf("Output file system:\n")
		dumpFileSystem(targetFS, "/")
		assertFileContents(t, "flag", targetFS, "/var/lib/output")
	})

	t.Run("Able to load a remote git repo", func(t *testing.T) {
		sourceFS := memfs.New()
		targetFS := memfs.New()
		cfg := File{
			GitModule: []GitModule{
				{
					Name:       "subject-under-test",
					Repository: "https://github.com/meschbach/empire-test.git",
					Ref:        "main",
					WorkingDir: "ctf",
				},
			},
		}
		env := NewModuleEnv(sourceFS, "/etc/empire", targetFS, "/", nil)
		require.NoError(t, env.resolve(&cfg))

		fmt.Printf("Output file system:\n")
		dumpFileSystem(targetFS, "/")
		assertFileContents(t, "flag", targetFS, "/var/lib/qa/flag")
	})
}

func listDir(t *testing.T, fs billy.Filesystem, path string) {
	ls, err := fs.ReadDir(path)
	require.NoError(t, err)
	for _, fi := range ls {
		if fi.IsDir() {
			fmt.Printf("%s/\n", fs.Join(path, fi.Name()))
			listDir(t, fs, fs.Join(path, fi.Name()))
		} else {
			fmt.Printf("%s\n", fs.Join(path, fi.Name()))
		}
	}
}

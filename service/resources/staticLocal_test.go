package resources

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"path/filepath"
	"testing"
)

func TestStaticLocalFile(t *testing.T) {
	fs := memfs.New()

	t.Run("When parent directory does not exists and file is relative to the module", func(t *testing.T) {
		moduleBase := "/etc/empire"
		targetFile := "/etc/systemd/system/consul.service"
		targetFileContents := "test"
		sourceFile := "consul.service"

		resolvedSourceFile := filepath.Join(moduleBase, sourceFile)
		writeTestFile(t, fs, resolvedSourceFile, targetFileContents)

		cfg := File{
			StaticLocalFiles: []StaticLocalFile{
				{
					Target: targetFile,
					Source: sourceFile,
				},
			},
		}
		env := NewModuleEnv(fs, "/etc/empire", fs, "/", nil)
		require.NoError(t, env.resolve(&cfg))
		assertFileContents(t, targetFileContents, fs, targetFile)
	})
}

func writeTestFile(t *testing.T, in billy.Filesystem, named string, content string) {
	file, err := in.Create(named)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, file.Close())
	}()
	_, err = io.WriteString(file, content)
	require.NoError(t, err)
}

func assertFileContents(t *testing.T, expected string, fs billy.Filesystem, fileName string) {
	t.Helper()
	info, err := fs.Stat(fileName)
	require.NoError(t, err, "stat error on %q", fileName)
	if assert.False(t, info.IsDir(), "not a directory") {
		return
	}

	handle, err := fs.Open(fileName)
	require.NoError(t, err)
	defer func() { require.NoError(t, handle.Close()) }()
	actual, err := io.ReadAll(handle)
	require.NoError(t, err)
	assert.Equal(t, expected, string(actual))
}

package resources

import (
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDirectory(t *testing.T) {
	sourceFS := memfs.New()
	destinationFS := memfs.New()

	t.Run("When creating a new directory", func(t *testing.T) {
		exampleDir := "/some/path"
		cfg := File{
			Directory: []Directory{
				{
					Name: "test-dir",
					Path: exampleDir,
				},
			},
		}
		env := NewModuleEnv(sourceFS, "/etc/empire", destinationFS, "/", nil)
		require.NoError(t, env.resolve(&cfg))
		info, err := destinationFS.Stat(exampleDir)
		require.NoError(t, err, exampleDir)
		assert.True(t, info.IsDir())

		t.Run("When run again it is still a directory", func(t *testing.T) {
			env := NewModuleEnv(sourceFS, "/etc/empire", destinationFS, "/", nil)
			require.NoError(t, env.resolve(&cfg))
			info, err := destinationFS.Stat(exampleDir)
			require.NoError(t, err)
			assert.True(t, info.IsDir())
		})
	})
}

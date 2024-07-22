package resources

import (
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPackageState(t *testing.T) {
	cfgFS := memfs.New()
	vmFS := memfs.New()

	t.Run("When a module calls for a package", func(t *testing.T) {
		cfg := File{
			Package: []Package{
				{Name: "unbound"},
			},
		}
		virtualPackages := &CapturingPackageManager{}
		env := NewModuleEnv(cfgFS, "/etc/empire", vmFS, "/", virtualPackages)
		require.NoError(t, env.resolve(&cfg))

		if assert.Len(t, virtualPackages.InstallPackages, 1) {
			assert.Equal(t, "unbound", virtualPackages.InstallPackages[0])
		}
	})
}

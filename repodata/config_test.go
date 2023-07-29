package repodata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
	yamlContent := `
timeout_seconds: 123
listen: localhost:54321
filtered_repodata_dir: repodata-cache/filtered-x
channels:
  conda-forge:
    subdirs: [linux-64, noarch]
    allowlist_file: /test/conda-forge-allowlist.txt
  test:
    subdirs: [osx-64]
`
	tmpdir := t.TempDir()
	configFile := filepath.Join(tmpdir, "test.yaml")
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	c, err := LoadCondaRepoConfig(configFile)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assert.Equal(t, "https://conda.anaconda.org", c.CondaHost)
	assert.Equal(t, 123, c.TimeoutSeconds)
	assert.Equal(t, 1440, c.MaxAgeMinutes)
	assert.Equal(t, "localhost:54321", c.Listen)
	assert.Equal(t, 1440, c.CacheControlMaxAgeMinutes)
	assert.Equal(t, "repodata-cache/original", c.OriginalRepodataDir)
	assert.Equal(t, "repodata-cache/filtered-x", c.FilteredRepodataDir)

	assert.Equal(t, 2, len(c.Channels))
	assert.Equal(t, c.Channels["conda-forge"].Subdirs, []string{"linux-64", "noarch"})
	assert.Equal(t, c.Channels["conda-forge"].AllowlistFile, "/test/conda-forge-allowlist.txt")
	assert.Equal(t, c.Channels["test"].Subdirs, []string{"osx-64"})
	assert.Equal(t, c.Channels["test"].AllowlistFile, "")
}

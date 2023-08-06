package repodata

import (
	"os"

	"gopkg.in/yaml.v3"
)

type condaChannelConfig struct {
	Subdirs             []string `yaml:"subdirs"`
	AllowlistFile       string   `yaml:"allowlist_file"`
	RecurseDependencies bool     `yaml:"recurse_dependencies"`
}

type CondaRepoConfig struct {
	CondaHost                 string                        `yaml:"conda_host"`
	TimeoutSeconds            int                           `yaml:"timeout_seconds"`
	MaxAgeMinutes             int                           `yaml:"max_age_minutes"`
	Listen                    string                        `yaml:"listen"`
	CacheControlMaxAgeMinutes int                           `yaml:"cache_control_max_age_minutes"`
	OriginalRepodataDir       string                        `yaml:"original_repodata_dir"`
	FilteredRepodataDir       string                        `yaml:"filtered_repodata_dir"`
	Channels                  map[string]condaChannelConfig `yaml:"channels"`
}

// LoadCondaRepoConfig loads a configuration file and returns the config
func LoadCondaRepoConfig(filename string) (*CondaRepoConfig, error) {
	c := CondaRepoConfig{
		CondaHost:                 "https://conda.anaconda.org",
		TimeoutSeconds:            120,
		MaxAgeMinutes:             1440,
		Listen:                    "localhost:8080",
		CacheControlMaxAgeMinutes: 1440,
		OriginalRepodataDir:       "repodata-cache/original",
		FilteredRepodataDir:       "repodata-cache/filtered",
		Channels:                  make(map[string]condaChannelConfig),
	}
	err := c.Load(filename)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *CondaRepoConfig) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, c)
	return err
}

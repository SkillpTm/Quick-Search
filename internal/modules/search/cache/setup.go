// Package cache handles everything that has to do with the generation of the cache for the Search function, to the generation of our folder structure and importing of the config.
package cache

import (
	"fmt"
	"os"
	"strings"

	"github.com/skillptm/Bolt/internal/util"
)

// ConfigJSONData is made to structer and order the data for the config.json
type ConfigJSONData struct {
	MaxCPUThreadPercentage float64  `json:"maxCPUThreadPercentage"`
	DefaultDirs            []string `json:"defaultDirs"`
	ExtendedDirs           []string `json:"extendedDirs"`
	ExcludeFromDefaultDirs Rules    `json:"excludeFromDefaultDirs"`
	ExcludeDirs            Rules    `json:"excludeDirs"`
}

// Rules is made to structer and order the data for the config.json
type Rules struct {
	Name  []string `json:"name"`
	Path  []string `json:"path"`
	Regex []string `json:"regex"`
}

// setupFolders validates all files/folders we need to exist
func setupFolders() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("setupFolders: couldn't access the user's cache dir:\n--> %w", err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("setupFolders: couldn't access the user's config dir:\n--> %w", err)
	}

	err = validateFolders([]string{
		fmt.Sprintf("%s/bolt/", cacheDir),
		fmt.Sprintf("%s/bolt/", configDir),
	})
	if err != nil {
		return fmt.Errorf("setupFolders: couldn't validate default folders:\n--> %w", err)
	}

	err = validateFiles([]string{
		fmt.Sprintf("%s/bolt/search-cache.json", cacheDir),
		fmt.Sprintf("%s/bolt/config.json", configDir),
	})
	if err != nil {
		return fmt.Errorf("setupFolders: couldn't validate default files:\n--> %w", err)
	}

	return nil
}

// validateFolders checks, if our folders in the user's config/cache dirs exists
func validateFolders(dirsToCheck []string) error {
	for _, dir := range dirsToCheck {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			continue
		}

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("validateFolders: failed to create directory at %s:\n--> %w", dir, err)
		}
	}

	return nil
}

// validateFiles checks, if config.json and search-cache.json exist
func validateFiles(filesToCheck []string) error {
	for _, file := range filesToCheck {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			continue
		}

		_, err := os.Create(file)
		if err != nil {
			return fmt.Errorf("validateFiles: failed to create file at %s:\n--> %w", file, err)
		}

		if strings.HasSuffix(file, "config.json") {
			resetConfig()
		}
	}

	return nil
}

// resetConfig resets the config file to the default settings
func resetConfig() error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resetConfig: couldn't find the user's home dir:\n--> %w", err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("resetConfig: couldn't access the user's config dir:\n--> %w", err)
	}

	defaultConfig := ConfigJSONData{
		MaxCPUThreadPercentage: 0.25, // percentage of threads that may be used, always rounding the threads up
		DefaultDirs: []string{
			fmt.Sprintf("%s/", homedir),
		},
		ExtendedDirs: []string{
			"/",
		},
		ExcludeFromDefaultDirs: Rules{
			Name: []string{},
			Path: []string{},
			Regex: []string{
				fmt.Sprintf(`^%s/\.[^/]+/?$`, homedir),
			},
		},
		ExcludeDirs: Rules{
			Name: []string{
				".git",
				"node_modules",
				"steamapps",
			},
			Path:  []string{},
			Regex: []string{},
		},
	}

	err = util.OverwriteJSON(fmt.Sprintf("%s/bolt/config.json", configDir), defaultConfig)
	if err != nil {
		return fmt.Errorf("resetConfig: couldn't reset default config:\n--> %w", err)
	}

	return nil
}

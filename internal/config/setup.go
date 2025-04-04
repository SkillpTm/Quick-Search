// Package config validates our local folder structure and to retrieve the config data
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/skillptm/Bolt/internal/util"
)

// Setup validates all files/folders we need to exist and returns their paths
func setup() ([]string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("setup: couldn't access the user's cache dir:\n--> %w", err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("setup: couldn't access the user's config dir:\n--> %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("setup: couldn't access the user's config dir:\n--> %w", err)
	}

	err = validateFolders([]string{
		fmt.Sprintf("%s/Bolt/", cacheDir),
		fmt.Sprintf("%s/Bolt/", configDir),
		fmt.Sprintf("%s/.local/share/Bolt/", homeDir),
	})
	if err != nil {
		return nil, fmt.Errorf("setup: couldn't validate default folders:\n--> %w", err)
	}

	files := []string{
		fmt.Sprintf("%s/Bolt/default_cache.json", cacheDir),
		fmt.Sprintf("%s/Bolt/extended_cache.json", cacheDir),
		fmt.Sprintf("%s/Bolt/config.json", configDir),
		fmt.Sprintf("%s/.local/share/Bolt/error.log", homeDir),
	}

	err = validateFiles(files)
	if err != nil {
		return nil, fmt.Errorf("setup: couldn't validate default files:\n--> %w", err)
	}

	return files, nil
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

// validateFiles checks, if config.json and search_cache.json exist
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

	defaultConfig := Config{
		MaxCPUThreadPercentage:      0.25, // percentage of threads that may be used, always rounding the threads up
		ShortCutEnd:                 "space",
		DefaultDirsCacheUpdateTime:  30,  // in seconds
		ExtendedDirsCacheUpdateTime: 600, // in seconds
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

	err = util.OverwriteJSON(fmt.Sprintf("%s/Bolt/config.json", configDir), true, defaultConfig)
	if err != nil {
		return fmt.Errorf("resetConfig: couldn't reset default config:\n--> %w", err)
	}

	return nil
}

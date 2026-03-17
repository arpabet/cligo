/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package cligo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.arpabet.com/glue"
)

// resolveConfigFiles finds the first existing file from the given paths
// and returns glue beans for property loading.
// For .properties/.yaml/.yml/.json files, returns a glue.PropertySource bean.
// For .env files, returns a parsed glue.MapPropertySource bean.
func resolveConfigFiles(paths []string) ([]interface{}, error) {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".properties", ".yaml", ".yml", ".json", ".toml":
			return []interface{}{&glue.PropertySource{File: "file:" + path}}, nil
		default:
			return nil, fmt.Errorf("unsupported config file format: %s", ext)
		}
	}
	return nil, nil // no file found, not an error
}

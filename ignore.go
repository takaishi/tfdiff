package tfdiff

import (
	"path"
	"path/filepath"
	"strings"
)

type ParseOptions struct {
	IgnoreFiles []string
}

func loadIgnorePatterns(extra []string) []string {
	patterns := make([]string, 0, len(extra))
	for _, pattern := range extra {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		patterns = append(patterns, pattern)
	}
	return patterns
}

func shouldIgnore(relPath string, baseName string, patterns []string) (bool, error) {
	relPath = filepath.ToSlash(relPath)

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		pattern = filepath.ToSlash(pattern)
		pattern = strings.TrimPrefix(pattern, "/")
		pattern = strings.TrimSuffix(pattern, "/")

		if strings.Contains(pattern, "/") {
			matched, err := path.Match(pattern, relPath)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
			continue
		}

		matched, err := path.Match(pattern, baseName)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

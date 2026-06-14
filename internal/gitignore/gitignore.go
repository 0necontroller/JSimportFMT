package gitignore

import (
	"path/filepath"
	"strings"

	"github.com/sabhiram/go-gitignore"
)

var defaultIgnoredDirs = []string{
	"node_modules",
	"dist",
	"build",
	".next",
	".nuxt",
	"coverage",
	"out",
}

type Matcher struct {
	ignore      *ignore.GitIgnore
	allowedDirs map[string]bool
}

// NewMatcher creates a new gitignore matcher.
func NewMatcher(gitignorePath string, allowDirs []string) (*Matcher, error) {
	var ig *ignore.GitIgnore
	var err error

	if gitignorePath != "" {
		ig, err = ignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			// If file doesn't exist, we can still use default ignore rules.
			ig = ignore.CompileIgnoreLines()
		}
	} else {
		ig = ignore.CompileIgnoreLines()
	}

	allowed := make(map[string]bool)
	for _, dir := range allowDirs {
		allowed[dir] = true
	}

	return &Matcher{
		ignore:      ig,
		allowedDirs: allowed,
	}, nil
}

// IsIgnored checks if a given file or directory path is ignored.
func (m *Matcher) IsIgnored(path string, isDir bool) bool {
	// First check explicit allow overrides for directories.
	if isDir {
		base := filepath.Base(path)
		if m.allowedDirs[base] {
			return false
		}
	} else {
		// For files, if their parent dir is allowed, we don't automatically allow the file,
		// but we already traversed into the directory.
	}

	// Check default ignored directories (if they are a directory, or part of the path).
	// We can just check if any part of the path matches the default ignored dirs.
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		for _, ignoredDir := range defaultIgnoredDirs {
			if part == ignoredDir && !m.allowedDirs[ignoredDir] {
				return true
			}
		}
	}

	if m.ignore != nil {
		if m.ignore.MatchesPath(path) {
			return true
		}
	}

	return false
}

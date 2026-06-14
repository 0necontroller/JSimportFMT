package gitignore

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sabhiram/go-gitignore"
)

var defaultIgnoredDirs = []string{
	// Next.js & React Ecosystem
	".next", ".cache", "dist", "build", ".expo",
	// Vue, Nuxt, & Svelte Ecosystem
	".nuxt", ".output", ".svelte-kit",
	// Angular & Other Meta-Frameworks
	".angular", ".astro", ".remix", "out",
	// Modern JS Bundlers & Runtimes
	".rsbuild", ".parcel-cache", ".turbo", ".vercel", ".netlify",
	// JavaScript/TypeScript
	"node_modules", "bower_components", "jspm_packages", "typings",
	// Python
	".venv", "venv", "env", "conda-meta", "__pycache__",
	// Java/Kotlin/Scala
	".gradle", "target", ".idea",
	// C#/.NET
	"bin", "obj", "packages", ".vs",
	// Go, Ruby
	"vendor", ".bundle",
	// General Version Control & OS
	".git", ".svn", ".hg", "CVS", ".Trash", "Thumbs.db",
}

type Matcher struct {
	ignore       *ignore.GitIgnore
	allowedDirs  map[string]bool
	ignoreLines  []string
}

// NewMatcher creates a new gitignore matcher.
func NewMatcher(gitignorePath string, allowDirs []string) (*Matcher, error) {
	var ig *ignore.GitIgnore
	var err error
	var lines []string

	if gitignorePath != "" {
		ig, err = ignore.CompileIgnoreFile(gitignorePath)
		if err == nil {
			// Read the lines to store them
			content, err := os.ReadFile(gitignorePath)
			if err == nil {
				for _, line := range strings.Split(string(content), "\n") {
					line = strings.TrimSpace(line)
					if line != "" && !strings.HasPrefix(line, "#") {
						lines = append(lines, line)
					}
				}
			}
		} else {
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
		ignoreLines: lines,
	}, nil
}

func (m *Matcher) GetDefaultIgnored() []string {
	return defaultIgnoredDirs
}

func (m *Matcher) GetGitignoreLines() []string {
	return m.ignoreLines
}

// IsIgnored checks if a given file or directory path is ignored.
func (m *Matcher) IsIgnored(path string, isDir bool) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")

	// First check explicit allow overrides for directories.
	for _, part := range parts {
		if m.allowedDirs[part] {
			return false
		}
	}

	// Check default ignored directories
	for _, part := range parts {
		for _, ignoredDir := range defaultIgnoredDirs {
			if part == ignoredDir {
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

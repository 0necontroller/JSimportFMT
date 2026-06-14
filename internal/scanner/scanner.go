package scanner

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/0necontroller/jsimportfmt/internal/gitignore"
)

// ValidExtensions lists the supported file extensions.
var ValidExtensions = map[string]bool{
	".js":  true,
	".jsx": true,
	".mjs": true,
	".cjs": true,
	".ts":  true,
	".tsx": true,
	".mts": true,
	".cts": true,
}

// FindProjectRoot searches upwards for a .git directory.
// It limits the search to a maximum of 2 parent directories.
// If none is found, it falls back to the target's directory.
func FindProjectRoot(target string) string {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return filepath.Dir(target)
	}

	dir := absTarget
	info, err := os.Stat(dir)
	if err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	startDir := dir

	// Limit search to 2 levels up
	for level := 0; level <= 2; level++ {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return startDir
}

func GetMatcher(target string, allowDirs []string) (*gitignore.Matcher, error) {
	root := FindProjectRoot(target)
	return gitignore.NewMatcher(filepath.Join(root, ".gitignore"), allowDirs)
}

func Scan(target string, allowDirs []string, fileChan chan<- string, errChan chan<- error) {
	defer close(fileChan)

	root := FindProjectRoot(target)

	matcher, err := gitignore.NewMatcher(filepath.Join(root, ".gitignore"), allowDirs)
	if err != nil {
		errChan <- err
		return
	}

	info, err := os.Stat(target)
	if err != nil {
		errChan <- err
		return
	}

	if !info.IsDir() {
		if ValidExtensions[filepath.Ext(target)] {
			fileChan <- target
		}
		return
	}

	err = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" || relPath == "." {
			relPath = path
		}

		isDir := d.IsDir()

		if path == root || path == target {
			// Don't ignore the root directory being scanned
		} else if isDir && d.Name() == ".git" {
			return fs.SkipDir
		} else if matcher.IsIgnored(relPath, isDir) {
			if isDir {
				return fs.SkipDir
			}
			return nil
		}

		if !isDir {
			ext := filepath.Ext(d.Name())
			if ValidExtensions[ext] {
				fileChan <- path
			}
		}

		return nil
	})

	if err != nil {
		errChan <- err
	}
}

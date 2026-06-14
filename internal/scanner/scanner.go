package scanner

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/0necontroller/jsimportfmt/internal/gitignore"
)

var ErrNotInGitRepo = errors.New("target is not inside a git repository")

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

func ValidateGitRepo(target string) (string, error) {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	dir := absTarget
	info, err := os.Stat(dir)
	if err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return "", ErrNotInGitRepo
}

func GetMatcher(target string, allowDirs []string) (*gitignore.Matcher, error) {
	gitRoot, err := ValidateGitRepo(target)
	if err != nil {
		return nil, err
	}
	return gitignore.NewMatcher(filepath.Join(gitRoot, ".gitignore"), allowDirs)
}

func Scan(target string, allowDirs []string, fileChan chan<- string, errChan chan<- error) {
	defer close(fileChan)

	gitRoot, err := ValidateGitRepo(target)
	if err != nil {
		errChan <- err
		return
	}

	matcher, err := gitignore.NewMatcher(filepath.Join(gitRoot, ".gitignore"), allowDirs)
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

		relPath, _ := filepath.Rel(gitRoot, path)
		if relPath == "" || relPath == "." {
			relPath = path
		}

		isDir := d.IsDir()

		if path == gitRoot || path == target {
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

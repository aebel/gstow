package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Linker struct {
	Simulate bool
	Verbose  bool
}

func NewLinker() *Linker {
	return &Linker{}
}

func (l *Linker) log(format string, args ...interface{}) {
	if l.Verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func (l *Linker) CreateSymlink(source, target string) error {
	l.log("LINK: %s -> %s", target, source)

	if l.Simulate {
		return nil
	}

	targetDir := filepath.Dir(target)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("target already exists: %s", target)
	}

	relSource, err := filepath.Rel(targetDir, source)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	return os.Symlink(relSource, target)
}

func (l *Linker) RemoveSymlink(target string) error {
	l.log("UNLINK: %s", target)

	if l.Simulate {
		return nil
	}

	fi, err := os.Lstat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat %s: %w", target, err)
	}

	if fi.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("not a symlink: %s", target)
	}

	return os.Remove(target)
}

func (l *Linker) IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

func (l *Linker) ReadLink(path string) (string, error) {
	return os.Readlink(path)
}

func (l *Linker) PointsTo(symlink, target string) (bool, error) {
	dest, err := os.Readlink(symlink)
	if err != nil {
		return false, err
	}

	symlinkDir := filepath.Dir(symlink)
	resolved := filepath.Join(symlinkDir, dest)
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return false, err
	}

	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}

	return resolved == targetAbs, nil
}

func (l *Linker) IsOwnedByStow(target, stowDir string) (bool, error) {
	if !l.IsSymlink(target) {
		return false, nil
	}

	dest, err := l.ReadLink(target)
	if err != nil {
		return false, err
	}

	targetDir := filepath.Dir(target)
	resolved := filepath.Join(targetDir, dest)
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return false, err
	}

	stowDirAbs, err := filepath.Abs(stowDir)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(resolved, stowDirAbs), nil
}

func (l *Linker) RemoveEmptyDirs(path string) error {
	if l.Simulate {
		return nil
	}

	for {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		if len(entries) > 0 {
			return nil
		}

		l.log("RMDIR: %s", path)

		if err := os.Remove(path); err != nil {
			return err
		}

		path = filepath.Dir(path)
	}
}

package stow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aebel/gstow/internal/config"
	"github.com/aebel/gstow/internal/link"
)

type Stow struct {
	cfg     *config.Config
	linker  *link.Linker
	stowDir string
}

func New(stowDir string, cfg *config.Config) *Stow {
	return &Stow{
		cfg:     cfg,
		linker:  link.NewLinker(),
		stowDir: stowDir,
	}
}

func (s *Stow) SetSimulate(simulate bool) {
	s.linker.Simulate = simulate
}

func (s *Stow) SetVerbose(verbose bool) {
	s.linker.Verbose = verbose
}

func (s *Stow) Stow(packages ...string) error {
	for _, pkg := range packages {
		if err := s.stowPackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stow) Unstow(packages ...string) error {
	for _, pkg := range packages {
		if err := s.unstowPackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

func (s *Stow) Restow(packages ...string) error {
	if err := s.Unstow(packages...); err != nil {
		return err
	}
	return s.Stow(packages...)
}

func (s *Stow) getPackagePath(pkg string) string {
	return filepath.Join(s.stowDir, pkg)
}

func (s *Stow) getPackageConfig(pkg string) (*config.Config, error) {
	cfg, err := config.LoadConfig(s.stowDir, pkg)
	if err != nil {
		return nil, err
	}

	if s.cfg.Target != "" {
		cfg.Target = s.cfg.Target
	}
	if s.cfg.Dir != "" {
		cfg.Dir = s.cfg.Dir
	}
	cfg.Ignore = append(cfg.Ignore, s.cfg.Ignore...)

	return cfg, nil
}

func (s *Stow) stowPackage(pkg string) error {
	pkgPath := s.getPackagePath(pkg)

	fi, err := os.Stat(pkgPath)
	if err != nil {
		return fmt.Errorf("package %s not found: %w", pkg, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("package %s is not a directory", pkg)
	}

	cfg, err := s.getPackageConfig(pkg)
	if err != nil {
		return err
	}

	if cfg.Target == "" && len(cfg.Paths) == 0 {
		return fmt.Errorf("no target directory specified")
	}

	return s.processPackageDir(pkgPath, pkg, cfg)
}

func (s *Stow) processPackageDir(pkgPath, pkg string, cfg *config.Config) error {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if cfg.ShouldIgnore(name) {
			continue
		}

		sourcePath := filepath.Join(pkgPath, name)

		var targetPath string
		if mappedTarget, ok := cfg.Paths[name]; ok {
			targetPath = mappedTarget
		} else {
			targetDir := cfg.Target
			if !cfg.Flatten {
				targetDir = filepath.Join(cfg.Target, pkg)
			}
			targetPath = filepath.Join(targetDir, name)
		}

		if entry.IsDir() {
			if err := s.stowDirEntry(sourcePath, targetPath, cfg); err != nil {
				return err
			}
		} else {
			if err := s.stowFileEntry(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Stow) processDir(sourceDir, targetDir string, cfg *config.Config) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if cfg.ShouldIgnore(name) {
			continue
		}

		sourcePath := filepath.Join(sourceDir, name)
		targetPath := filepath.Join(targetDir, name)

		if entry.IsDir() {
			if err := s.stowDirEntry(sourcePath, targetPath, cfg); err != nil {
				return err
			}
		} else {
			if err := s.stowFileEntry(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Stow) stowDirEntry(sourcePath, targetPath string, cfg *config.Config) error {
	targetFi, err := os.Lstat(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return s.linker.CreateSymlink(sourcePath, targetPath)
	}

	if targetFi.Mode()&os.ModeSymlink != 0 {
		owned, err := s.linker.IsOwnedByStow(targetPath, s.stowDir)
		if err != nil {
			return err
		}
		if owned {
			if err := s.unfoldTree(targetPath); err != nil {
				return err
			}
			return s.processDir(sourcePath, targetPath, cfg)
		}
		return fmt.Errorf("conflict: %s is a symlink not owned by gstow", targetPath)
	}

	if targetFi.IsDir() {
		return s.processDir(sourcePath, targetPath, cfg)
	}

	return fmt.Errorf("conflict: %s exists and is not a directory", targetPath)
}

func (s *Stow) stowFileEntry(sourcePath, targetPath string) error {
	targetFi, err := os.Lstat(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return s.linker.CreateSymlink(sourcePath, targetPath)
	}

	if targetFi.Mode()&os.ModeSymlink != 0 {
		owned, err := s.linker.IsOwnedByStow(targetPath, s.stowDir)
		if err != nil {
			return err
		}
		if owned {
			return nil
		}
		return fmt.Errorf("conflict: %s is a symlink not owned by gstow", targetPath)
	}

	return fmt.Errorf("conflict: %s already exists", targetPath)
}

func (s *Stow) unfoldTree(targetPath string) error {
	dest, err := s.linker.ReadLink(targetPath)
	if err != nil {
		return err
	}

	targetDir := filepath.Dir(targetPath)
	sourcePath := filepath.Join(targetDir, dest)
	sourcePath, err = filepath.Abs(sourcePath)
	if err != nil {
		return err
	}

	if err := s.linker.RemoveSymlink(targetPath); err != nil {
		return err
	}

	if err := os.Mkdir(targetPath, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		newSource := filepath.Join(sourcePath, name)
		newTarget := filepath.Join(targetPath, name)

		if entry.IsDir() {
			if err := s.linker.CreateSymlink(newSource, newTarget); err != nil {
				return err
			}
		} else {
			if err := s.linker.CreateSymlink(newSource, newTarget); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Stow) unstowPackage(pkg string) error {
	pkgPath := s.getPackagePath(pkg)

	cfg, err := s.getPackageConfig(pkg)
	if err != nil {
		return err
	}

	if cfg.Target == "" && len(cfg.Paths) == 0 {
		return fmt.Errorf("no target directory specified")
	}

	return s.unstowPackageDir(pkgPath, pkg, cfg)
}

func (s *Stow) unstowPackageDir(pkgPath, pkg string, cfg *config.Config) error {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if cfg.ShouldIgnore(name) {
			continue
		}

		sourcePath := filepath.Join(pkgPath, name)

		var targetPath string
		if mappedTarget, ok := cfg.Paths[name]; ok {
			targetPath = mappedTarget
		} else {
			targetDir := cfg.Target
			if !cfg.Flatten {
				targetDir = filepath.Join(cfg.Target, pkg)
			}
			targetPath = filepath.Join(targetDir, name)
		}

		if entry.IsDir() {
			if err := s.unstowDirEntry(sourcePath, targetPath, cfg); err != nil {
				return err
			}
		} else {
			if err := s.unstowFileEntry(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Stow) unstowDir(sourceDir, targetDir string, cfg *config.Config) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if cfg.ShouldIgnore(name) {
			continue
		}

		sourcePath := filepath.Join(sourceDir, name)
		targetPath := filepath.Join(targetDir, name)

		if entry.IsDir() {
			if err := s.unstowDirEntry(sourcePath, targetPath, cfg); err != nil {
				return err
			}
		} else {
			if err := s.unstowFileEntry(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Stow) unstowDirEntry(sourcePath, targetPath string, cfg *config.Config) error {
	targetFi, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if targetFi.Mode()&os.ModeSymlink != 0 {
		owned, err := s.linker.IsOwnedByStow(targetPath, s.stowDir)
		if err != nil {
			return err
		}
		if !owned {
			return nil
		}

		pointsTo, err := s.linker.PointsTo(targetPath, sourcePath)
		if err != nil {
			return err
		}
		if pointsTo {
			return s.linker.RemoveSymlink(targetPath)
		}
		return nil
	}

	if targetFi.IsDir() {
		if err := s.unstowDir(sourcePath, targetPath, cfg); err != nil {
			return err
		}

		entries, err := os.ReadDir(targetPath)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			s.linker.RemoveEmptyDirs(targetPath)
		}

		return nil
	}

	return nil
}

func (s *Stow) unstowFileEntry(sourcePath, targetPath string) error {
	if !s.linker.IsSymlink(targetPath) {
		return nil
	}

	owned, err := s.linker.IsOwnedByStow(targetPath, s.stowDir)
	if err != nil {
		return err
	}
	if !owned {
		return nil
	}

	pointsTo, err := s.linker.PointsTo(targetPath, sourcePath)
	if err != nil {
		return err
	}
	if pointsTo {
		return s.linker.RemoveSymlink(targetPath)
	}

	return nil
}

func (s *Stow) ListPackages() ([]string, error) {
	entries, err := os.ReadDir(s.stowDir)
	if err != nil {
		return nil, err
	}

	var packages []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			packages = append(packages, entry.Name())
		}
	}
	return packages, nil
}

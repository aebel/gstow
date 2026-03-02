package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var DefaultIgnorePatterns = []string{
	`^\.git$`,
	`^\.gitignore$`,
	`^\.stowrc$`,
	`^CVS$`,
	`^\.#`,
	`.*\.swp$`,
	`.*~$`,
}

type Config struct {
	Target  string
	Dir     string
	Ignore  []string
	Flatten bool
	Paths   map[string]string
}

type RcFile struct {
	Default  *Config
	Packages map[string]*Config
}

func New() *Config {
	return &Config{
		Ignore: append([]string{}, DefaultIgnorePatterns...),
		Paths:  make(map[string]string),
	}
}

func (c *Config) Merge(other *Config) {
	if other.Target != "" {
		c.Target = other.Target
	}
	if other.Dir != "" {
		c.Dir = other.Dir
	}
	c.Ignore = append(c.Ignore, other.Ignore...)
	if other.Flatten {
		c.Flatten = true
	}
	for k, v := range other.Paths {
		c.Paths[k] = v
	}
}

func (c *Config) Clone() *Config {
	paths := make(map[string]string)
	for k, v := range c.Paths {
		paths[k] = v
	}
	return &Config{
		Target:  c.Target,
		Dir:     c.Dir,
		Ignore:  append([]string{}, c.Ignore...),
		Flatten: c.Flatten,
		Paths:   paths,
	}
}

func NewRcFile() *RcFile {
	return &RcFile{
		Default:  New(),
		Packages: make(map[string]*Config),
	}
}

func (r *RcFile) GetConfig(pkg string) *Config {
	cfg := r.Default.Clone()
	if pkgCfg, ok := r.Packages[pkg]; ok {
		cfg.Merge(pkgCfg)
	}
	return cfg
}

func (r *RcFile) LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open config file %s: %w", path, err)
	}
	defer file.Close()

	var currentPkg string
	var inPathsSection bool
	var currentConfig *Config = r.Default

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := line[1 : len(line)-1]
			if sectionName == "" || strings.Contains(sectionName, "=") {
				continue
			}

			if strings.HasSuffix(sectionName, ".paths") {
				pkgName := strings.TrimSuffix(sectionName, ".paths")
				inPathsSection = true
				currentPkg = pkgName
				if _, ok := r.Packages[currentPkg]; !ok {
					r.Packages[currentPkg] = &Config{Ignore: []string{}, Paths: make(map[string]string)}
				}
				currentConfig = r.Packages[currentPkg]
			} else {
				inPathsSection = false
				currentPkg = sectionName
				if currentPkg != "" {
					if _, ok := r.Packages[currentPkg]; !ok {
						r.Packages[currentPkg] = &Config{Ignore: []string{}, Paths: make(map[string]string)}
					}
					currentConfig = r.Packages[currentPkg]
				} else {
					currentConfig = r.Default
				}
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = expandEnv(value)

		if inPathsSection {
			currentConfig.Paths[key] = value
			continue
		}

		switch key {
		case "target":
			currentConfig.Target = value
		case "dir":
			currentConfig.Dir = value
		case "ignore":
			currentConfig.Ignore = append(currentConfig.Ignore, value)
		case "flatten":
			currentConfig.Flatten = value == "true" || value == "1" || value == "yes"
		}
	}

	return scanner.Err()
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "~") {
		home, _ := os.UserHomeDir()
		s = home + s[1:]
	}

	result := os.ExpandEnv(s)
	return result
}

func LoadConfig(stowDir, pkg string) (*Config, error) {
	rc := NewRcFile()

	homeRc := filepath.Join(os.Getenv("HOME"), ".stowrc")
	_ = rc.LoadFile(homeRc)

	_ = rc.LoadFile(filepath.Join(stowDir, ".stowrc"))

	cfg := rc.GetConfig(pkg)

	if cfg.Dir == "" {
		cfg.Dir = stowDir
	}

	return cfg, nil
}

func (c *Config) ShouldIgnore(path string) bool {
	for _, pattern := range c.Ignore {
		matched, err := regexp.MatchString(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}

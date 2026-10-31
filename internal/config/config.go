// Package config loads the optional .migrate-lanes.toml from the repo root.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// FileName is the config file looked up at the repository root.
const FileName = ".migrate-lanes.toml"

// Path pins a migration directory, optionally forcing the tool.
type Path struct {
	Dir  string `toml:"dir"`
	Tool string `toml:"tool"`
}

// Config is the whole file. Every field is optional.
//
//	base = "main"
//	remotes = false
//	ignore_branches = ["dependabot/*"]
//	[[paths]]
//	dir = "db/migrations"
//	tool = "golang-migrate"
type Config struct {
	Base           string   `toml:"base"`
	Remotes        bool     `toml:"remotes"`
	IgnoreBranches []string `toml:"ignore_branches"`
	Paths          []Path   `toml:"paths"`
}

// Load reads FileName from root. A missing file yields an empty config.
func Load(root string, validTools []string) (Config, error) {
	var c Config
	md, err := toml.DecodeFile(filepath.Join(root, FileName), &c)
	if errors.Is(err, fs.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return c, fmt.Errorf("%s: %w", FileName, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		return c, fmt.Errorf("%s: unknown key %q", FileName, undecoded[0].String())
	}
	return c, c.validate(validTools)
}

// validate rejects paths escaping the repo and unknown tools.
func (c *Config) validate(validTools []string) error {
	for i, p := range c.Paths {
		clean := path.Clean(strings.TrimSuffix(filepath.ToSlash(p.Dir), "/"))
		if p.Dir == "" || clean == "." || path.IsAbs(clean) || strings.HasPrefix(clean, "../") || clean == ".." {
			return fmt.Errorf("%s: paths[%d].dir %q must be a relative directory inside the repo", FileName, i, p.Dir)
		}
		c.Paths[i].Dir = clean
		if p.Tool != "" && !contains(validTools, p.Tool) {
			return fmt.Errorf("%s: paths[%d].tool %q unknown (supported: %s)", FileName, i, p.Tool, strings.Join(validTools, ", "))
		}
	}
	for _, g := range c.IgnoreBranches {
		if _, err := path.Match(g, ""); err != nil {
			return fmt.Errorf("%s: ignore_branches pattern %q: %w", FileName, g, err)
		}
	}
	return nil
}

// Ignored reports whether a branch name matches an ignore_branches glob.
func (c Config) Ignored(branch string) bool {
	for _, g := range c.IgnoreBranches {
		if ok, _ := path.Match(g, branch); ok {
			return true
		}
	}
	return false
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

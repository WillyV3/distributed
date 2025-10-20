package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the distributed config
type Config struct {
	Groups map[string][]string `yaml:"groups"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Groups: map[string][]string{
			"dev": {},
		},
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "distributed", "config.yaml"), nil
}

// Load reads the configuration file
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, create default
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration file
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetGroup returns hosts in a group
func (c *Config) GetGroup(name string) ([]string, error) {
	hosts, ok := c.Groups[name]
	if !ok {
		return nil, fmt.Errorf("group %q not found", name)
	}
	return hosts, nil
}

// AddToGroup adds a host to a group
func (c *Config) AddToGroup(group, host string) {
	if c.Groups == nil {
		c.Groups = make(map[string][]string)
	}
	c.Groups[group] = append(c.Groups[group], host)
}

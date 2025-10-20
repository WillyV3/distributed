package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// SSHHost represents a host from SSH config
type SSHHost struct {
	Alias    string
	Hostname string
	User     string
	Port     string
}

// ParseSSHConfig reads ~/.ssh/config and extracts host configurations
func ParseSSHConfig() ([]SSHHost, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []SSHHost
	var currentHost *SSHHost

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.ToLower(fields[0])
		value := fields[1]

		switch key {
		case "host":
			// Skip wildcard hosts
			if strings.Contains(value, "*") {
				currentHost = nil
				continue
			}

			currentHost = &SSHHost{
				Alias: value,
				Port:  "22", // default
			}
			hosts = append(hosts, *currentHost)

		case "hostname":
			if currentHost != nil {
				hosts[len(hosts)-1].Hostname = value
			}

		case "user":
			if currentHost != nil {
				hosts[len(hosts)-1].User = value
			}

		case "port":
			if currentHost != nil {
				hosts[len(hosts)-1].Port = value
			}
		}
	}

	return hosts, scanner.Err()
}

// GetHost finds a host by alias
func GetHost(hosts []SSHHost, alias string) *SSHHost {
	for _, h := range hosts {
		if h.Alias == alias {
			return &h
		}
	}
	return nil
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSSHConfig(t *testing.T) {
	// Create temporary SSH config for testing
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(sshDir, "config")
	configContent := `# Test SSH config
Host homelab
    HostName 100.72.192.70
    User wv3
    Port 22

Host sonia-mac
    HostName 100.75.170.108
    User sonia

# This should be ignored
Host *
    ServerAliveInterval 60

Host basic
    HostName example.com

Host with-custom-port
    HostName custom.example.com
    User admin
    Port 2222
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Temporarily override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	hosts, err := ParseSSHConfig()
	if err != nil {
		t.Fatalf("ParseSSHConfig failed: %v", err)
	}

	// Verify we got the expected hosts (excluding wildcard)
	expectedHosts := map[string]SSHHost{
		"homelab": {
			Alias:    "homelab",
			Hostname: "100.72.192.70",
			User:     "wv3",
			Port:     "22",
		},
		"sonia-mac": {
			Alias:    "sonia-mac",
			Hostname: "100.75.170.108",
			User:     "sonia",
			Port:     "22", // default
		},
		"basic": {
			Alias:    "basic",
			Hostname: "example.com",
			User:     "",
			Port:     "22",
		},
		"with-custom-port": {
			Alias:    "with-custom-port",
			Hostname: "custom.example.com",
			User:     "admin",
			Port:     "2222",
		},
	}

	if len(hosts) != len(expectedHosts) {
		t.Errorf("Expected %d hosts, got %d", len(expectedHosts), len(hosts))
	}

	for _, host := range hosts {
		expected, ok := expectedHosts[host.Alias]
		if !ok {
			t.Errorf("Unexpected host: %s", host.Alias)
			continue
		}

		if host.Hostname != expected.Hostname {
			t.Errorf("Host %s: expected hostname %s, got %s",
				host.Alias, expected.Hostname, host.Hostname)
		}

		if host.User != expected.User {
			t.Errorf("Host %s: expected user %s, got %s",
				host.Alias, expected.User, host.User)
		}

		if host.Port != expected.Port {
			t.Errorf("Host %s: expected port %s, got %s",
				host.Alias, expected.Port, host.Port)
		}
	}
}

func TestGetHost(t *testing.T) {
	hosts := []SSHHost{
		{Alias: "host1", Hostname: "192.168.1.1", User: "user1", Port: "22"},
		{Alias: "host2", Hostname: "192.168.1.2", User: "user2", Port: "2222"},
		{Alias: "host3", Hostname: "192.168.1.3", User: "user3", Port: "22"},
	}

	tests := []struct {
		name      string
		alias     string
		wantFound bool
		wantHost  *SSHHost
	}{
		{
			name:      "find existing host",
			alias:     "host2",
			wantFound: true,
			wantHost:  &SSHHost{Alias: "host2", Hostname: "192.168.1.2", User: "user2", Port: "2222"},
		},
		{
			name:      "host not found",
			alias:     "nonexistent",
			wantFound: false,
			wantHost:  nil,
		},
		{
			name:      "first host",
			alias:     "host1",
			wantFound: true,
			wantHost:  &SSHHost{Alias: "host1", Hostname: "192.168.1.1", User: "user1", Port: "22"},
		},
		{
			name:      "last host",
			alias:     "host3",
			wantFound: true,
			wantHost:  &SSHHost{Alias: "host3", Hostname: "192.168.1.3", User: "user3", Port: "22"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetHost(hosts, tt.alias)

			if tt.wantFound {
				if got == nil {
					t.Fatalf("Expected to find host %s, got nil", tt.alias)
				}

				if got.Alias != tt.wantHost.Alias ||
					got.Hostname != tt.wantHost.Hostname ||
					got.User != tt.wantHost.User ||
					got.Port != tt.wantHost.Port {
					t.Errorf("GetHost() = %+v, want %+v", got, tt.wantHost)
				}
			} else {
				if got != nil {
					t.Errorf("Expected nil for host %s, got %+v", tt.alias, got)
				}
			}
		})
	}
}

func TestParseSSHConfig_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	hosts, err := ParseSSHConfig()
	if err != nil {
		t.Fatalf("ParseSSHConfig failed on empty file: %v", err)
	}

	if len(hosts) != 0 {
		t.Errorf("Expected 0 hosts from empty config, got %d", len(hosts))
	}
}

func TestParseSSHConfig_CommentsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(sshDir, "config")
	configContent := `# Only comments
# No actual hosts
# Just testing
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	hosts, err := ParseSSHConfig()
	if err != nil {
		t.Fatalf("ParseSSHConfig failed: %v", err)
	}

	if len(hosts) != 0 {
		t.Errorf("Expected 0 hosts from comments-only config, got %d", len(hosts))
	}
}

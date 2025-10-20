package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultExcludes(t *testing.T) {
	// Verify critical excludes are present
	requiredExcludes := []string{
		".git",
		"node_modules",
		".DS_Store",
		"__pycache__",
		"dist",
		"build",
	}

	for _, required := range requiredExcludes {
		found := false
		for _, exclude := range defaultExcludes {
			if exclude == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required exclude pattern '%s' not found in defaultExcludes", required)
		}
	}
}

func TestPathResolution(t *testing.T) {
	// Test converting absolute paths to relative paths from home
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		inputPath  string
		wantPrefix string
	}{
		{
			name:       "home directory path",
			inputPath:  filepath.Join(home, "projects", "myapp"),
			wantPrefix: "~/projects/myapp",
		},
		{
			name:       "direct home",
			inputPath:  home,
			wantPrefix: "~",
		},
		{
			name:       "nested path",
			inputPath:  filepath.Join(home, "dev", "go", "src", "app"),
			wantPrefix: "~/dev/go/src/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the path conversion logic from Push
			remotePath := strings.Replace(tt.inputPath, home, "~", 1)

			if remotePath != tt.wantPrefix {
				t.Errorf("Expected remote path '%s', got '%s'", tt.wantPrefix, remotePath)
			}
		})
	}
}

func TestRsyncArgsConstruction(t *testing.T) {
	tests := []struct {
		name            string
		dryRun          bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:   "normal sync",
			dryRun: false,
			wantContains: []string{
				"-avz",
				"--progress",
				"--exclude",
			},
			wantNotContains: []string{
				"--dry-run",
			},
		},
		{
			name:   "dry run sync",
			dryRun: true,
			wantContains: []string{
				"-avz",
				"--progress",
				"--exclude",
				"--dry-run",
			},
			wantNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build args like Push does
			args := []string{
				"-avz",
				"--progress",
			}

			if tt.dryRun {
				args = append(args, "--dry-run")
			}

			// Add excludes
			for _, exclude := range defaultExcludes {
				args = append(args, "--exclude", exclude)
			}

			argsStr := strings.Join(args, " ")

			for _, want := range tt.wantContains {
				if !strings.Contains(argsStr, want) {
					t.Errorf("Expected args to contain '%s', args: %s", want, argsStr)
				}
			}

			for _, wantNot := range tt.wantNotContains {
				if strings.Contains(argsStr, wantNot) {
					t.Errorf("Expected args to NOT contain '%s', args: %s", wantNot, argsStr)
				}
			}
		})
	}
}

func TestExcludePatternsFormat(t *testing.T) {
	// Verify exclude patterns build correctly for rsync
	var args []string

	for _, exclude := range defaultExcludes {
		args = append(args, "--exclude", exclude)
	}

	// Should have pairs of --exclude followed by pattern
	if len(args)%2 != 0 {
		t.Error("Exclude args should be in pairs of --exclude and pattern")
	}

	for i := 0; i < len(args); i += 2 {
		if args[i] != "--exclude" {
			t.Errorf("Expected --exclude at position %d, got %s", i, args[i])
		}
		if args[i+1] == "" {
			t.Errorf("Exclude pattern at position %d is empty", i+1)
		}
	}
}

func TestSyncPathConstruction(t *testing.T) {
	// Test the path construction for rsync
	tests := []struct {
		name       string
		absPath    string
		host       string
		remotePath string
		wantSource string
		wantDest   string
	}{
		{
			name:       "basic sync",
			absPath:    "/Users/test/project",
			host:       "homelab",
			remotePath: "~/project",
			wantSource: "/Users/test/project/",
			wantDest:   "homelab:~/project/",
		},
		{
			name:       "nested path",
			absPath:    "/Users/test/dev/app",
			host:       "server",
			remotePath: "~/dev/app",
			wantSource: "/Users/test/dev/app/",
			wantDest:   "server:~/dev/app/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate path construction from Push
			source := tt.absPath + "/"
			dest := tt.host + ":" + tt.remotePath + "/"

			if source != tt.wantSource {
				t.Errorf("Expected source '%s', got '%s'", tt.wantSource, source)
			}

			if dest != tt.wantDest {
				t.Errorf("Expected dest '%s', got '%s'", tt.wantDest, dest)
			}
		})
	}
}

func TestPullPathConstruction(t *testing.T) {
	// Test Pull path construction
	tests := []struct {
		name       string
		host       string
		remotePath string
		localPath  string
		wantSource string
		wantDest   string
	}{
		{
			name:       "pull from remote",
			host:       "homelab",
			remotePath: "~/project",
			localPath:  "/Users/test/project",
			wantSource: "homelab:~/project/",
			wantDest:   "/Users/test/project/",
		},
		{
			name:       "pull nested",
			host:       "server",
			remotePath: "/opt/app",
			localPath:  "/Users/test/app",
			wantSource: "server:/opt/app/",
			wantDest:   "/Users/test/app/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate path construction from Pull
			source := tt.host + ":" + tt.remotePath + "/"
			dest := tt.localPath + "/"

			if source != tt.wantSource {
				t.Errorf("Expected source '%s', got '%s'", tt.wantSource, source)
			}

			if dest != tt.wantDest {
				t.Errorf("Expected dest '%s', got '%s'", tt.wantDest, dest)
			}
		})
	}
}

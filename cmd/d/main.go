package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/WillyV3/distributed/internal/config"
	"github.com/WillyV3/distributed/internal/host"
	"github.com/WillyV3/distributed/internal/run"
	"github.com/WillyV3/distributed/internal/sync"
	"github.com/WillyV3/distributed/internal/ui"
	"github.com/spf13/cobra"
)

var (
	groupFlag  string
	hostFlag   string
	allFlag    bool
	dryRunFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "d",
		Short: "Distributed development across machines",
		Long:  "Manage distributed development across multiple machines using SSH and rsync",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&groupFlag, "group", "g", "dev", "Target group")
	rootCmd.PersistentFlags().StringVar(&hostFlag, "host", "", "Target specific host")
	rootCmd.PersistentFlags().BoolVar(&allFlag, "all", false, "Target all hosts in group")

	// Commands
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(loadCmd())
	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(tmuxCmd())
	rootCmd.AddCommand(configCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of all hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			var hosts []config.SSHHost
			var err error

			err = ui.Spin("Checking hosts", func() error {
				hosts, err = config.ParseSSHConfig()
				return err
			})

			if err != nil {
				return fmt.Errorf("failed to parse SSH config: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "HOST\tSTATUS\tADDRESS")

			for _, h := range hosts {
				status := "✓ online"
				if !host.CheckReachable(h.Alias, 2*time.Second) {
					status = "✗ offline"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", h.Alias, status, h.Hostname)
			}

			return w.Flush()
		},
	}
}

func loadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "load",
		Short: "Show load across all hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			hosts, err := getTargetHosts()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "HOST\tLOAD\tCPUS\tCPU%\tMEM%\tSCORE")

			var best *host.LoadInfo
			for _, h := range hosts {
				var info *host.LoadInfo
				err := ui.Spin(fmt.Sprintf("Checking %s", h), func() error {
					var loadErr error
					info, loadErr = host.GetLoad(h)
					return loadErr
				})

				if err != nil || info == nil || !info.Reachable {
					fmt.Fprintf(w, "%s\t-\t-\t-\t-\t-\n", h)
					continue
				}

				marker := ""
				if best == nil || info.Score < best.Score {
					best = info
					marker = " ←"
				}

				fmt.Fprintf(w, "%s\t%.2f\t%d\t%d%%\t%d%%\t%.2f%s\n",
					info.Host, info.Load, info.CPUs, info.CPUPct, info.MemPct, info.Score, marker)
			}

			return w.Flush()
		},
	}
}

func syncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [path]",
		Short: "Sync directory to remote hosts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			hosts, err := getTargetHosts()
			if err != nil {
				return err
			}

			if dryRunFlag {
				ui.Info("Dry run - no files will be transferred")
			}

			err = sync.Push(path, hosts, dryRunFlag)
			if err == nil {
				ui.Success("Sync complete")
			}
			return err
		},
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be synced")
	return cmd
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [command...]",
		Short: "Run command on best host",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.Join(args, " ")

			if allFlag {
				hosts, err := getTargetHosts()
				if err != nil {
					return err
				}
				ui.Info(fmt.Sprintf("Running on all hosts: %s", strings.Join(hosts, ", ")))
				return run.OnAll(hosts, command)
			}

			// Run on best host
			hosts, err := getTargetHosts()
			if err != nil {
				return err
			}

			var best *host.LoadInfo
			err = ui.Spin("Finding best host", func() error {
				var findErr error
				best, findErr = host.FindBest(hosts)
				return findErr
			})

			if err != nil {
				return err
			}

			ui.Info(fmt.Sprintf("Running on %s (score: %.2f)", best.Host, best.Score))
			return run.OnHost(best.Host, command)
		},
	}
}

func tmuxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tmux [host]",
		Short: "Sync current dir and attach to tmux",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetHost := args[0]

			// Sync current directory
			ui.Info("Syncing current directory")
			if err := sync.Push(".", []string{targetHost}, false); err != nil {
				return err
			}
			ui.Success("Sync complete")

			// Get current working directory (relative to home)
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			remotePath := strings.Replace(cwd, home, "~", 1)

			ui.Info(fmt.Sprintf("Connecting to %s", targetHost))

			// SSH and attach to tmux
			tmuxCmd := fmt.Sprintf("cd %s && (TERM=xterm-256color tmux attach-session -t dev 2>/dev/null || TERM=xterm-256color tmux new-session -s dev)", remotePath)

			sshCmd := exec.Command("ssh", "-t", targetHost, tmuxCmd)
			sshCmd.Stdin = os.Stdin
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr

			return sshCmd.Run()
		},
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse SSH config to get available hosts
			hosts, err := config.ParseSSHConfig()
			if err != nil {
				return err
			}

			// Create default config with all hosts in dev group
			cfg := config.DefaultConfig()
			for _, h := range hosts {
				cfg.AddToGroup("dev", h.Alias)
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			path, _ := config.ConfigPath()
			fmt.Printf("✓ Configuration created at %s\n", path)
			fmt.Printf("✓ Added %d hosts to 'dev' group\n", len(hosts))

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			fmt.Println("Groups:")
			for group, hosts := range cfg.Groups {
				fmt.Printf("  %s: %s\n", group, strings.Join(hosts, ", "))
			}

			return nil
		},
	})

	return cmd
}

// getTargetHosts returns the list of hosts to target based on flags
func getTargetHosts() ([]string, error) {
	// Specific host flag takes precedence
	if hostFlag != "" {
		return []string{hostFlag}, nil
	}

	// Load config and get group
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	hosts, err := cfg.GetGroup(groupFlag)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no hosts in group %q", groupFlag)
	}

	return hosts, nil
}

# distributed

Distribute workload across Tailscale-connected machines with intelligent load balancing.

## The Problem

Development on a MacBook Pro maxes out resources. I have a Linux homelab and another MacBook on my Tailscale network sitting idle. I needed a way to offload builds, tests, and heavy workloads to whichever machine has capacity.

## The Solution

CLI that reads your SSH config, checks machine load, and distributes work intelligently.

- Finds the least-loaded machine automatically
- Syncs files via rsync
- Runs commands over SSH
- Opens tmux sessions on remote machines

Pure Unix tools. No daemons, no APIs, no complexity.

## Installation

```bash
brew install willyv3/tap/distributed
```

Or from source:
```bash
git clone https://github.com/WillyV3/distributed.git
cd distributed
go install ./cmd/dw
```

## Usage

```bash
dw status              # Check which hosts are online
dw load                # Show CPU/memory load across hosts
dw sync .              # Sync current directory to remote hosts
dw run go build        # Run command on least-loaded machine
dw tmux homelab        # Open tmux session with synced files
```

## Configuration

Reads from `~/.ssh/config` automatically:

```
Host homelab
  HostName 100.72.192.70
  User wv3

Host build-server
  HostName 192.168.1.100
  User admin
```

Optional groups at `~/.config/distributed/config.yaml`:

```yaml
groups:
  dev:
    - homelab
    - build-server
```

## Commands

### dw status
Check reachability of all SSH hosts.

### dw load
Display load metrics. Score = (CPU% × 0.7) + (Memory% × 0.3). Lower is better.

### dw sync [path]
Sync directory to remote hosts using rsync.

Flags:
- `--dry-run` - Preview what would sync
- `--host <name>` - Target specific host
- `-g, --group <name>` - Target group (default: dev)

Auto-excludes: .git, node_modules, dist, build, .DS_Store, __pycache__

### dw run [command]
Execute command on best available machine or all machines.

Flags:
- `--all` - Run on all machines in parallel
- `--host <name>` - Target specific host
- `-g, --group <name>` - Target group

Examples:
```bash
dw run npm test                   # Runs on least-loaded machine
dw run --all "git pull"           # Runs on all machines
dw run --host homelab go build    # Runs on specific host
```

### dw tmux [host]
Sync current directory and attach to tmux session on remote host.

Example:
```bash
cd ~/projects/myapp
dw tmux homelab
# Now in tmux on homelab at ~/projects/myapp
# Files already synced, ready to work
```

## Examples

Heavy build:
```bash
cd ~/projects/myapp
dw sync .
dw run go build
```

Parallel testing:
```bash
dw sync ~/projects/myapp
dw run --all "cd ~/projects/myapp && go test ./..."
```

Remote development:
```bash
cd ~/projects/vinw
dw tmux homelab
```

## Requirements

Local machine:
- SSH client
- rsync
- tmux (optional, for tmux command)

Remote machines:
- SSH server
- rsync
- tmux (optional)

## Security

- No arbitrary code execution
- Uses SSH authentication from ~/.ssh/config
- No exposed services, everything over SSH
- No secrets in code, uses SSH agent

## License

MIT

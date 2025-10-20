# d - Distributed Development CLI

Pure Unix. One binary. No bullshit.

Manage distributed development across multiple machines using SSH, rsync, and tmux.

## Philosophy

- **Use existing tools**: SSH, rsync, tmux - no reinventing
- **Read from standards**: Parses `~/.ssh/config` automatically
- **Minimal dependencies**: Just Go stdlib + cobra + yaml
- **Secure**: Never executes untrusted code, uses SSH properly
- **Simple**: One command, clear subcommands

## Installation

```bash
cd ~/distributed
go install ./cmd/d
```

**Note**: If you have a shell function `d()`, you can:
- Use full path: `~/go/bin/d`
- Create alias: `alias dist='~/go/bin/d'`
- Remove the function from your shell config

## Quick Start

### 1. Initialize Config
```bash
d config init
```

This reads `~/.ssh/config` and creates `~/.config/distributed/config.yaml` with all hosts in a "dev" group.

### 2. Check Status
```bash
d status
```

Shows which machines are online.

### 3. Check Load
```bash
d load
```

Shows CPU, memory, and load scores. Lower score = better for running tasks.

### 4. Sync Files
```bash
# Sync current directory to all machines in dev group
d sync .

# Sync to specific host
d sync ~/projects/myapp --host homelab

# Dry run (see what would sync)
d sync . --dry-run
```

### 5. Run Commands
```bash
# Run on best (least loaded) machine
d run go build

# Run on all machines
d run --all "uname -a"

# Run on specific host
d run --host homelab "docker ps"
```

### 6. Tmux Integration
```bash
# Sync current directory and attach to tmux on homelab
d tmux homelab

# Inside tmux, you're in the same directory path
# Files are already synced
# Work as if local
```

## Commands

### `d status`
Show status of all hosts from SSH config.

### `d load`
Display load metrics across all hosts. Arrow (←) indicates best machine.

### `d sync [path]`
Sync directory to remote hosts using rsync.

**Flags:**
- `--dry-run` - Show what would be synced
- `--host <name>` - Sync to specific host
- `-g, --group <name>` - Sync to specific group (default: "dev")

**Auto-excludes:**
- `.git`, `node_modules`, `.DS_Store`
- `*.pyc`, `__pycache__`, `.venv`, `venv`
- `dist`, `build`, `.next`, `target`, `.terraform`

### `d run [command]`
Execute command on best available machine (or all machines).

**Flags:**
- `--all` - Run on all machines in parallel
- `--host <name>` - Run on specific host
- `-g, --group <name>` - Target specific group

**Examples:**
```bash
d run npm test                    # Runs on least-loaded machine
d run --all "git pull"           # Runs on all machines
d run --host homelab go build    # Runs on homelab specifically
```

### `d tmux [host]`
Sync current directory and attach to tmux session.

**What it does:**
1. Syncs current directory to remote host
2. SSHes to host
3. CDs to same directory path
4. Attaches to tmux (or creates new session)

**Example:**
```bash
# You're in ~/projects/vinw on Mac
pwd  # /Users/you/projects/vinw

d tmux homelab

# Now in tmux on homelab, at ~/projects/vinw
# Files are synced
# Ready to build/test
```

### `d config`
Manage configuration.

**Subcommands:**
- `d config init` - Create default config from SSH config
- `d config show` - Display current configuration

## Configuration

### SSH Config (`~/.ssh/config`)
`d` reads this automatically. Example:

```ssh-config
Host homelab
  HostName 100.72.192.70
  User wv3

Host build-server
  HostName build.example.com
  User deploy
  Port 2222
```

### d Config (`~/.config/distributed/config.yaml`)
Groups for targeting multiple hosts:

```yaml
groups:
  dev:
    - homelab
    - build-server
  production:
    - prod-web-01
    - prod-web-02
```

Use groups with `-g, --group` flag:
```bash
d sync . --group production
d run --group production "systemctl restart app"
```

## Examples

### Heavy Go Build
```bash
cd ~/projects/myapp
d sync .
d run go build -o myapp
```

`d` automatically picks the least-loaded machine.

### Parallel Testing
```bash
d sync ~/projects/myapp
d run --all "cd ~/projects/myapp && go test ./..."
```

Runs tests on all machines simultaneously.

### Remote Development
```bash
cd ~/projects/vinw
d tmux homelab

# Now in tmux on homelab
go run .
# Edit locally, files sync automatically if using watch mode
```

### Development with Watch Mode
```bash
# Terminal 1: Manual sync when needed
d sync .

# Terminal 2: Work on homelab
d tmux homelab
```

Or use `fswatch` for auto-sync:
```bash
# Terminal 1: Auto-sync on changes
fswatch -o . | xargs -n1 -I{} d sync .

# Terminal 2: Work remotely
d tmux homelab
```

## Security

- **No arbitrary code execution**: Commands are passed directly to SSH
- **Uses SSH authentication**: Respects `~/.ssh/config` and keys
- **No exposed services**: Everything over SSH tunnel
- **Input validation**: Host names and paths validated before use
- **No secrets in code**: Uses SSH agent for key management

## Dependencies

**Required on local machine:**
- Go 1.21+ (for building)
- SSH client
- rsync

**Optional:**
- tmux (for `d tmux` command)
- fswatch (for auto-sync workflows)

**Go dependencies:**
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML config parsing

## Workflows

### Workflow: Release Build
```bash
cd ~/projects/myapp
d sync .
d load  # Check which machine is best
d run goreleaser build --snapshot
```

### Workflow: Distributed Testing
```bash
d sync ~/projects/myapp
d run --all "cd ~/projects/myapp && make test"
```

### Workflow: Remote Dev Session
```bash
cd ~/projects/vinw
d tmux homelab
# Now in tmux on homelab, same directory
# Detach: Ctrl+B, D
# Reattach later: d tmux homelab
```

## Troubleshooting

### `d: command not found`
Binary is in `~/go/bin/d`. Either:
- Add `~/go/bin` to PATH: `export PATH=$PATH:~/go/bin`
- Use full path: `~/go/bin/d`
- Create alias: `alias dist='~/go/bin/d'`

### Shell function `d` conflicts
You might have a `d()` function (dir stack). Either:
- Rename shell function
- Use full path to binary
- Create alias with different name

### Host unreachable
- Check SSH config: `cat ~/.ssh/config`
- Test SSH manually: `ssh <host>`
- Verify Tailscale: `tailscale status`

### Sync fails
- Ensure rsync is installed on both machines
- Check file permissions
- Try dry-run first: `d sync . --dry-run`

## Why This vs Scripts?

**Old way (bash scripts):**
- Multiple commands: `dsync`, `drun`, `dtmux`
- Shell scripts, not portable
- Dependencies on `sshsync`, custom tools
- Hardcoded host names

**New way (d CLI):**
- One command: `d`
- Go binary, runs anywhere
- Uses standard tools: SSH, rsync
- Reads from `~/.ssh/config`

## Contributing

This is a personal tool but feel free to fork and modify.

Principles:
- Keep it simple
- Use Unix tools
- No unnecessary dependencies
- Secure by default

## License

MIT

---

**Built with Unix philosophy: Do one thing well.**

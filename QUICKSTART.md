# Quick Start Guide

## Setup Complete! 🎉

Your distributed development environment is ready to use with your homelab. Sonia's MacBook needs one more step (see below).

## What's Working Now

✅ **homelab** (Linux) - Fully configured and accessible
✅ **dsync** - File synchronization tool
✅ **drun** - Intelligent load-balanced execution
✅ **dtmux** - Tmux session management

⏳ **sonia-mac** - Needs SSH key (see Final Setup Step below)

## Try It Now

### 1. Check Machine Load

```bash
cd ~/distributed
./drun --show-load
```

This shows CPU, memory, and load across all machines, highlighting the best one for work.

### 2. Run a Command on Best Machine

```bash
./drun "uname -a && uptime"
```

This automatically picks the least-loaded machine and runs the command there.

### 3. Sync Files

```bash
# Push current directory to all machines
./dsync push .

# Check sync status
./dsync status
```

### 4. Use with Tmux

```bash
# Sync current directory and attach to homelab tmux
./dtmux sync homelab

# Inside tmux, work normally:
# - Code, run commands, etc.
# - Detach with: Ctrl+B, then D
# - Reattach later: ./dtmux attach homelab
```

## Final Setup Step: Enable Sonia's MacBook

To use Sonia's MacBook, add your SSH key to it:

**On Sonia's MacBook, open Terminal and run:**

```bash
mkdir -p ~/.ssh
echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOijiPwMW0oXdrdXOF1J4ljVg9a5v7i5KD5mxtyLLnU1 williamvansickleiii@Williams-MacBook-Pro.local" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

**Then test from your Mac:**

```bash
ssh sonia-mac "echo 'Connected!'"
```

Once this works, all three tools will work with Sonia's Mac too!

## Make It Even Easier

Add these aliases to your `~/.zshrc` or `~/.bashrc`:

```bash
# Source the aliases file
source ~/distributed/shell-aliases.sh
```

Then reload your shell:
```bash
source ~/.zshrc  # or source ~/.bashrc
```

Now you can use short commands:
```bash
ds push .           # dsync push .
dr npm run build    # drun npm run build
dt sync homelab     # dtmux sync homelab
dstat              # Check machine load
dhome              # Sync and attach to homelab
```

## Common Workflows

### Heavy Build/Compile
```bash
# Sync code to all machines
ds push .

# Run build on best available machine
dr npm run build
```

### Remote Development
```bash
# Sync project and jump into tmux on homelab
dt sync homelab

# Work in tmux...
# When done: Ctrl+B, D (detach)
# Later: dt attach homelab (reattach)
```

### Watch and Auto-Sync
```bash
# Terminal 1: Watch for changes
ds watch ~/projects/myapp

# Terminal 2: Work remotely
dt sync homelab

# Files auto-sync as you save locally!
```

### Check Before Running
```bash
# See machine load
dr --show-load

# Test where command would run
dr --dry-run "npm run build"
```

## What Each Tool Does

### dsync - File Synchronization
- `dsync push <path>` - Push files to all dev machines
- `dsync pull <host> <path>` - Pull files from a machine
- `dsync status` - Show machine status
- `dsync watch <path>` - Auto-sync on file changes

### drun - Smart Execution
- `drun <command>` - Run on least-loaded machine
- `drun --all <command>` - Run on all machines
- `drun --host <name> <command>` - Run on specific machine
- `drun --show-load` - Display load metrics

### dtmux - Tmux Management
- `dtmux sync <host>` - Sync code and attach to tmux
- `dtmux attach <host>` - Attach to existing session
- `dtmux list` - List all tmux sessions
- `dtmux monitor` - Multi-machine monitoring view

## Getting Help

Each tool has built-in help:
```bash
./dsync --help
./drun --help
./dtmux --help
```

For detailed information, see `README.md`.

## Troubleshooting

**Machine not responding?**
```bash
tailscale status
ssh homelab "echo OK"
```

**Sync not working?**
```bash
./dsync status
```

**Want to see what would happen?**
```bash
./dsync push . --dry-run
./drun --dry-run "npm run build"
```

---

You now have a personal compute cluster! Your three machines work together as if they're one. 🚀

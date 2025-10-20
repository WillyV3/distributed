#!/usr/bin/env bash
#
# Distributed Development Shell Aliases
# Source this file in your ~/.zshrc or ~/.bashrc:
#   source ~/distributed/shell-aliases.sh
#

# Get the directory where this script is located
DISTRIBUTED_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Main tool aliases
alias ds="${DISTRIBUTED_DIR}/dsync"
alias dr="${DISTRIBUTED_DIR}/drun"
alias dt="${DISTRIBUTED_DIR}/dtmux"

# Common operations
alias dpush='ds push .'
alias dpull='ds pull'
alias dstat='dr --show-load'
alias dlist='ds status'

# Execution shortcuts
alias dbuild='dr npm run build'
alias dtest='dr npm test'
alias dlint='dr npm run lint'
alias dall='dr --all'

# Tmux shortcuts
alias dhome='dt sync homelab'
alias dsonia='dt sync sonia-mac'
alias dmon='dt monitor'

# Status checks
alias dmachines='sshsync ls'
alias dgroups='sshsync ls'

# Sync shortcuts
alias dwatch='ds watch .'
alias dsync-here='ds push .'

# Helper function to run on best machine
dbest() {
    dr "$@"
}

# Helper function to run on all machines
deverywhere() {
    dr --all "$@"
}

# Helper function to sync and run
dsync-run() {
    local command="$1"
    ds push . && dr "$command"
}

echo "✓ Distributed development aliases loaded"
echo "  Tools: ds (sync), dr (run), dt (tmux)"
echo "  Try: dstat, dpush, dbuild, dhome, dmon"

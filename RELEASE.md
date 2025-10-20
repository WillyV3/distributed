# Release Process

## GitHub Repository

- **Repo**: https://github.com/WillyV3/distributed
- **Current Version**: v0.0.1

## Release Script

Location: `scripts/release.sh`

### Usage

```bash
./scripts/release.sh [major|minor|patch] "Optional commit message"
```

### Examples

```bash
# Patch release (bug fixes)
./scripts/release.sh patch "Fix sync error handling"

# Minor release (new features)
./scripts/release.sh minor "Add new sync options"

# Major release (breaking changes)
./scripts/release.sh major "Redesign CLI interface"
```

## What It Does

1. Commits any pending changes
2. Bumps version (major/minor/patch)
3. Builds binary to verify compilation
4. Creates and pushes git tag
5. Calculates SHA256 for tarball
6. Updates Homebrew formula (if exists)
7. Generates changelog from commits
8. Creates GitHub release

## Installation Methods

### From Source
```bash
git clone https://github.com/WillyV3/distributed.git
cd distributed
go install ./cmd/dw
```

### From Release
```bash
go install github.com/WillyV3/distributed/cmd/dw@latest
```

### From Specific Version
```bash
go install github.com/WillyV3/distributed/cmd/dw@v0.0.1
```

## First Release Created

- Version: v0.0.1
- Date: 2025-10-20
- URL: https://github.com/WillyV3/distributed/releases/tag/v0.0.1

## Next Steps

To create more releases:
```bash
cd ~/distributed
./scripts/release.sh patch "Description of changes"
```

The script handles everything automatically.

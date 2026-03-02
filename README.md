# gstow

A dotfile manager written in Go, inspired by [GNU Stow](https://www.gnu.org/software/stow/) with support for per-package configuration and path mappings.

## Features

- **Symlink farm management** - Creates and manages symlinks from packages to target directories
- **INI-style configuration** - Central `.stowrc` with per-package settings
- **Path mappings** - Map individual files to different target locations
- **Tree folding/unfolding** - Minimal symlinks with automatic directory structure management
- **Safe operations** - Never deletes files not owned by gstow
- **Ignore patterns** - Flexible file exclusion with regex support
- **Simulation mode** - Preview changes before applying

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap aebel/gstow
brew install gstow
```

### From Source

```bash
git clone https://github.com/aebel/gstow.git
cd gstow
go build -o build/gstow ./cmd/gstow
```

### Download Binary

Download the latest release for your platform from the [Releases](https://github.com/aebel/gstow/releases) page.

## Usage

### Basic Commands

```bash
gstow <package>        # Stow a package
gstow -D <package>     # Unstow a package
gstow -R <package>     # Restow (unstow then stow)
```

### Options

```
-d <dir>      Stow directory (default: current directory)
-t <dir>      Target directory (overrides .stowrc)
-n            Simulate; don't make any changes
-v            Verbose output
-h            Show help
```

### Examples

```bash
gstow nvim                    # Stow nvim package
gstow -D zsh                  # Unstow zsh package
gstow -R nvim                 # Restow nvim package
gstow -n -v nvim              # Preview stow operation
gstow -t ~/.config nvim       # Override target directory
```

## Configuration

Create a `.stowrc` file in your stow directory:

### Basic Configuration

```ini
# Default target for all packages
target=~/.config

# Package-specific configuration
[nvim]
# inherits ~/.config from default

[zsh]
target=~
flatten=true
```

### Configuration Options

| Option | Description |
|--------|-------------|
| `target` | Target directory for symlinks |
| `dir` | Stow directory (usually auto-detected) |
| `flatten` | Don't append package name to target (true/false) |
| `ignore` | Regex pattern for files to ignore (can be repeated) |

### Path Mappings

Map individual files to specific locations:

```ini
[zsh]
target=~/.config/zsh

[zsh.paths]
.zshenv = ~/.zshenv
```

This configuration:
- Links `.zshrc` to `~/.config/zsh/.zshrc`
- Links `.zshenv` to `~/.zshenv`

### Ignore Patterns

```ini
ignore=\.bak
ignore=\.orig
ignore=^README
```

Default ignore patterns:
- `.git`, `.gitignore`, `.stowrc`
- `CVS`, `.*.swp`, `*~`, `.#*`

## Example Setup

### Directory Structure

```
~/dotfiles/
├── .stowrc
├── nvim/
│   └── init.vim
├── zsh/
│   ├── .zshrc
│   └── .zshenv
└── git/
    └── .gitconfig
```

### .stowrc

```ini
target=~/.config

[nvim]
# → ~/.config/nvim/init.vim

[zsh]
target=~/.config/zsh

[zsh.paths]
.zshenv = ~/.zshenv
# → ~/.config/zsh/.zshrc
# → ~/.zshenv

[git]
target=~
flatten=true
# → ~/.gitconfig
```

### Stow All Packages

```bash
cd ~/dotfiles
gstow nvim zsh git
```

## How It Works

### Tree Folding

When possible, gstow creates a single symlink for an entire directory:

```
# Without folding: many symlinks
~/.config/nvim/init.vim -> ~/dotfiles/nvim/init.vim

# With folding: one symlink
~/.config/nvim -> ~/dotfiles/nvim
```

### Tree Unfolding

When multiple packages share a directory, gstow automatically unfolds:

```bash
gstow nvim     # Creates: ~/.config/nvim -> ~/dotfiles/nvim
gstow zsh      # Unfolds: ~/.config/nvim/, ~/.config/zsh/
```

### Ownership

gstow only manages symlinks it owns (pointing into the stow directory). It will never:
- Delete or modify files it doesn't own
- Override existing files (unless they're gstow-owned symlinks)

## Differences from GNU Stow

| Feature | GNU Stow | gstow |
|---------|----------|-------|
| Config location | `.stowrc` per directory | Single `.stowrc` with sections |
| Path mappings | No | Yes (per-file targets) |
| Package target | Parent directory | Configurable per package |
| Flatten option | No | Yes |

## License

MIT License - see [LICENSE](LICENSE) for details.

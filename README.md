# dotctl

`dotctl` manages dotfiles with a bare Git repository whose work tree is normally your home directory. It also manages executable dependency scripts called runnables.

> This is a personal project. Inspect commands and dependency scripts before using them against your home directory.

## Installation

```bash
go install github.com/scottjr632/dotctl@latest
```

## Initialize

Create a new bare repository:

```bash
dotctl init
```

Or clone an existing one:

```bash
dotctl init --clone https://github.com/you/dotfiles.git
```

By default, dotctl stores configuration in `~/.config/dotctl/config`, the bare repository in `~/.cfg/.dotfiles`, and uses your home directory as the Git work tree.

## Common commands

```bash
dotctl doctor
dotctl status
dotctl list
dotctl is-tracked .zshrc
dotctl track .zshrc
dotctl update
dotctl pull
dotctl push
dotctl dependencies list
```

Run `dotctl --help` or `dotctl COMMAND --help` for the complete command and flag list.

## Using dotctl from an agent or script

Start with the local, read-only health check:

```bash
dotctl doctor --json
```

Use `--non-interactive` to make dotctl fail instead of opening an editor or prompt. Preview mutations with `--dry-run`, then run the same command without it:

```bash
dotctl --non-interactive --dry-run track .zshrc --message "Track zsh config"
dotctl --non-interactive track .zshrc --message "Track zsh config"

dotctl --non-interactive --dry-run update --message "Update managed dotfiles"
dotctl --non-interactive update --message "Update managed dotfiles"
```

Commands that support structured inspection expose `--json`:

```bash
dotctl doctor --json
dotctl config show --json
dotctl status --json
dotctl list --json
dotctl is-tracked .zshrc --json
dotctl dependencies list --json
```

Successful JSON output has a stable envelope:

```json
{"ok":true,"data":{}}
```

Errors are written to stderr and return a nonzero exit status. Dotctl does not fetch from the network after unrelated commands; network access occurs only for explicit operations such as `init --clone`, `pull`, and `push`.

Use `--yes` with confirmed non-interactive operations, such as deleting or running all dependency scripts:

```bash
dotctl --non-interactive --yes dependencies delete brew
dotctl --non-interactive --yes dependencies all
```

Runnables execute local shell scripts with your user permissions. An agent should inspect a script before invoking `dependencies run` or `dependencies all`.

### Install the agent skill

Dotctl embeds an [Agent Skills](https://agentskills.io/) compatible skill. Install it into the shared user skill directory:

```bash
dotctl agent install-skill
```

This writes `~/.agents/skills/dotctl/SKILL.md`, which is discovered by pi and other compatible harnesses. Use `dotctl agent print-skill` to inspect it first, `--dry-run` to preview installation, or `--force` to replace an older copy. Reload or restart the agent after installation.

### Isolated operation

Override both state locations when an agent must not operate on the real home directory:

```bash
dotctl \
  --config-dir /tmp/dotctl/config \
  --work-tree /tmp/dotctl/home \
  --non-interactive \
  status --json
```

Environment equivalents are also supported:

```bash
export DOTCTL_CONFIG_DIR=/tmp/dotctl/config
export DOTCTL_WORK_TREE=/tmp/dotctl/home
```

## Development

Requires Go 1.21 or newer.

```bash
task check
```

Without Task:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

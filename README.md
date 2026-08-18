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

### Set up another computer

Clone the existing dotfiles repository, then check its tracked files out into your home directory:

```bash
dotctl init --clone https://github.com/you/dotfiles.git
dotctl checkout --backup-existing
```

`--backup-existing` moves files or symlinks that would block the first checkout to `~/.config/dotctl/backups/<timestamp>` before checking out the repository. It leaves unrelated files alone and prints the backup location. Use `--backup-dir <path>` to choose a different empty destination. Preview the operation before changing anything:

```bash
dotctl --json --dry-run checkout --backup-existing
dotctl --json checkout --backup-existing
```

If the home directory is already empty or contains no conflicting tracked paths, plain `dotctl checkout` is sufficient. The backup option is deliberately restricted to the first checkout so it cannot silently replace edits in an established work tree.

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

Sync before making changes on a computer, then commit and push those changes for the other computers:

```bash
dotctl pull
# Edit tracked dotfiles.
dotctl update --message "Update dotfiles"
dotctl push
```

`update` stages modifications and deletions to files that are already tracked. Use `track` for a new file. Dotctl uses Git's normal merge behavior and stops for manual conflict resolution rather than overwriting divergent edits.

Run `dotctl --help` or `dotctl COMMAND --help` for the complete command and flag list.

## Using dotctl from an agent or script

Start with the local, read-only health check:

```bash
dotctl doctor --json
```

Use `--non-interactive` to make dotctl fail instead of opening an editor or prompt. `--json` enables non-interactive behavior automatically. Preview mutations with `--dry-run`, then run the same command without it:

```bash
dotctl --json --dry-run track .zshrc --message "Track zsh config"
dotctl --json track .zshrc --message "Track zsh config"

dotctl --json --dry-run update --message "Update managed dotfiles"
dotctl --json update --message "Update managed dotfiles"
```

`--json` is global and works with inspection, planning, mutation, help, and error output:

```bash
dotctl --json --help
dotctl doctor --json
dotctl config show --json
dotctl status --json
dotctl list --json
dotctl is-tracked .zshrc --json
dotctl dependencies list --json
```

Every JSON invocation writes exactly one document to stdout. The `kind` is `result`, `plan`, or `error`:

```json
{"ok":true,"kind":"result","data":{}}
{"ok":true,"kind":"plan","data":{"actions":[{"operation":"stage","description":"Stage .zshrc in the dotfiles repository"}]}}
{"ok":false,"kind":"error","error":{"code":"CONFIG_NOT_FOUND","message":"dotctl config not found"}}
```

JSON errors return a nonzero exit status and keep stderr empty. Git and runnable output is captured under `data.output` instead of being mixed with the JSON document. Stable error codes are `CONFIG_NOT_FOUND`, `CONFIG_INVALID`, `INVALID_ARGUMENT`, `PERMISSION_DENIED`, `EXTERNAL_COMMAND_FAILED`, `DOCTOR_UNHEALTHY`, `JSON_UNSUPPORTED`, and the fallback `COMMAND_FAILED`.

Put global flags before `git` when using the passthrough command, for example `dotctl --json git status --short`.

Dotctl does not fetch from the network after unrelated commands; network access occurs only for explicit operations such as `init --clone`, `pull`, and `push`.

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

---
name: dotctl
description: Inspect, track, commit, synchronize, and check dotfiles managed by the dotctl CLI, select per-machine dotfile variants, and inspect or run configured dependency scripts. Use when the user asks about their managed dotfiles, dotfile repository status, shell/editor configuration tracked by dotctl, machine-specific dotfile variants, or dotctl runnables.
compatibility: Requires the dotctl and git executables.
---

# dotctl

Use dotctl through the shell. Prefer its non-interactive and structured interfaces instead of answering prompts or opening editors.

## Start safely

1. Confirm the executable exists with `command -v dotctl`.
2. Run `dotctl doctor --json`. This performs local checks without changing files or accessing the network.
3. Inspect state with `dotctl status --json` and `dotctl list --json`.
4. Before a mutation, run the exact command once with `--json --dry-run`.
5. Execute only after the structured plan matches the user's request, then verify with `dotctl status --json`.

Do not add `--config-dir` or `--work-tree` when operating on the user's real dotfiles unless the user provided those paths. Those flags select a different dotctl environment.

## Read-only commands

```bash
dotctl --json --help
dotctl doctor --json
dotctl config show --json
dotctl status --json
dotctl list --json
dotctl is-tracked <path> --json
dotctl profile show --json
dotctl profile list --json
dotctl dependencies list --json
```

Each JSON invocation writes exactly one document to stdout with `kind` set to `result`, `plan`, or `error`. Errors use `{"ok":false,"kind":"error","error":{"code":"...","message":"..."}}` and return a nonzero exit status; stderr remains empty. Git and runnable output appears in `data.output`. `is-tracked` returning `tracked: false` is a successful query, not a command failure.

## Mutating dotfiles

Use JSON mode and supply a commit message for unattended commits. JSON mode automatically disables prompts and editors:

```bash
dotctl --json --dry-run track <path> --message "Track <path>"
dotctl --json track <path> --message "Track <path>"

dotctl --json --dry-run update --message "Update managed dotfiles"
dotctl --json update --message "Update managed dotfiles"
```

Put global flags before the `git` passthrough command so dotctl can parse them: `dotctl --json git status --short`.

Use explicit `pull` and `push` commands for network synchronization. Ordinary inspection commands do not fetch.

For the first checkout after cloning a dotfiles repository, preview and back up paths that would otherwise block checkout:

```bash
dotctl --json --dry-run checkout --backup-existing
dotctl --json checkout --backup-existing
```

The result reports `backup_dir` and `backed_up`. The backup option is rejected after the first checkout; it is not a replacement for resolving ordinary Git changes or merge conflicts.

## Per-machine variants

A tracked file whose name contains `##` followed by comma-separated conditions is a variant, such as `.gitconfig##hostname.workbook` or `.zshrc##os.darwin,arch.arm64`. Conditions match on `hostname`, `os`, `arch`, and the configured `profile` name; every condition must match. `dotctl profile apply` symlinks each plain path to the most specific matching variant, and `checkout` applies variants automatically unless `--skip-profile` is passed.

```bash
dotctl profile show --json
dotctl profile list --json
dotctl --json --dry-run profile apply
dotctl --json profile apply
```

- Read `profile list --json` before creating a variant, so the new file name uses selectors this machine actually reports.
- Edit the variant file, not the linked path, when the user asks to change a machine-specific setting. Editing through the link works, but naming the variant makes the commit clear.
- A `PROFILE_CONFLICT` error means a target path already holds real content. Report the paths in `data.conflicts` and ask the user before using `--force`, which replaces those files without a backup.
- Never use `--force` to resolve a conflict whose reason mentions an already-tracked path; untracking the plain file is the only correct fix.
- `data.invalid` lists variants whose conditions could not be parsed. These are usually typos in a selector name and are worth reporting.

## Runnables

Runnables are executable local scripts and may install packages or change the system. Before running one:

1. Use `dotctl config show --json` to locate `dependencies_dir`.
2. Read the selected script.
3. Run `dotctl --json --dry-run dependencies run <name>`.
4. Execute it only when its contents and effects match the user's request.

Never run `dependencies all` merely to explore what it does. For an explicitly requested unattended operation, use both `--non-interactive` and `--yes`:

```bash
dotctl --json --dry-run dependencies all
dotctl --json --yes dependencies all
```

## Safety rules

- Do not edit the bare repository directly; use dotctl or Git through `dotctl git`.
- Do not treat printed output as success without checking the exit status.
- Do not use `--yes` to bypass ambiguity about what the user requested.
- Do not run dependency scripts before inspecting them.
- Do not use `profile apply --force` without the user's explicit agreement; it deletes the existing file at the target path.
- Use `--config-dir` and `--work-tree` together when deliberately testing in a sandbox.

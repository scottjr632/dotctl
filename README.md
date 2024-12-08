## dotctl

dotctl is a command line tool for managing dotfiles.

Note: this is mainly a personal project to help me manage my dotfiles :) 

## Installation

```bash
go install github.com/scottjr632/dotctl@latest
```

## Usage

```bash
dotctl [command]
```

### Commands

#### `init`

Initializes a new dotfile repository.

```bash
dotctl init
```

#### `add`

Adds a new dotfile to the repository.

```bash
dotctl add <dotfile>
```

#### `remove`

Removes a dotfile from the repository.

```bash
dotctl remove <dotfile>
```

#### `list`

Lists all dotfiles in the repository.

```bash
dotctl list
```

#### `update`

Updates all dotfiles in the repository.

```bash
dotctl update
```

#### `help`

Displays help information for a command.

```bash
dotctl help <command>
```

## License

MIT License

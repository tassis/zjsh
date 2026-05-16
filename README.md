# zjsh

`zjsh` aggregates configured projects, `zellij` sessions, and `zoxide` paths into one selector-friendly list, then connects the selected target through `zellij`.

It does not provide its own TUI. Pipe `zjsh list` into `fzf`, `gum`, or any selector that returns one selected line.

## What It Does

- Lists configured projects, live `zellij` sessions, resurrectable `zellij` sessions, and `zoxide` paths
- Emits plain selector labels by default, or JSON with `zjsh list --json`
- Accepts either raw targets or selector labels in `zjsh connect <target>`
- Creates new `zellij` sessions from project paths, startup commands, layouts, or layout files
- Can recreate a resurrected project session when `restart_on_resurrection` is enabled

## Dependencies

- `zellij`: used for session discovery, attach, switch, create, and delete operations
- `zoxide`: used as an additional path source
- `fzf` or `gum`: optional selector tools

`list` and `connect` use whatever sources are available. `doctor` is stricter and reports missing `zellij` or `zoxide` binaries as failures.

## Installation

```bash
go install github.com/saweima12/zjsh/cmd/zjsh@latest
```

Make sure your Go binary directory, usually `$GOBIN` or `$GOPATH/bin`, is in `PATH`.

## Quick Start

```bash
zjsh config init
zjsh doctor
zjsh list | fzf
zjsh connect "◆ api"
```

Selector symbols are accepted but not required:

```bash
zjsh connect api
zjsh connect "◆ api"
zjsh connect "● api"
zjsh connect "→ /Users/example/work/api"
```

## Commands

```bash
zjsh list
```

Print one selector label per line.

```bash
zjsh list --json
```

Print full entry metadata for scripting.

```bash
zjsh connect <target>
```

Resolve a project name, session name, or full path, then attach, switch, or create the matching `zellij` session.

```bash
zjsh doctor
```

Validate external binaries, config parsing, project paths, and layout files.

```bash
zjsh config init [--path <file>]
```

Create a sample config file. Existing files are not overwritten.

## Entry Labels

Plain `zjsh list` output is a single column designed for selectors:

- `◆ <name>`: configured project
- `● <name>`: live `zellij` session
- `↺ <name>`: resurrectable `zellij` session
- `→ <path>`: `zoxide` path

Configured projects are identified by project path and project name, so two configured projects are not merged just because they resolve to the same session basename. Other sources are merged when they refer to the same project path or session name. When sources merge, configured project fields are preferred while live session state from `zellij` is preserved. Any existing `zellij` session, including one already merged into a configured project, is preferred over a `zoxide` path with the same basename. Path-only `zoxide` entries do not merge with each other by basename, so different directories with the same basename stay separate.

## Connect Behavior

`zjsh connect <target>` strips known selector symbols first, so both `api` and `◆ api` resolve to the same target.

Resolution order is:

1. project name
2. session name
3. full path

When the target is a live session, `zjsh` attaches to it outside `zellij` or switches to it from inside `zellij`.

When the target is not a live session, `zjsh` creates or switches to a session using the resolved entry:

- `layout_file` uses the configured zellij layout file
- `layout` uses a named zellij layout
- `startup` creates a generated one-pane layout that runs the command through the configured shell
- path-only entries create a session in that directory

## Config

Default config path:

```text
~/.config/zjsh/config.kdl
```

Create it with:

```bash
zjsh config init
```

Example:

```kdl
defaults {
  shell "sh"
  restart_on_resurrection false
}

project "api" {
  path "/Users/example/work/api"
  session "api"
  startup "nvim ."
  restart_on_resurrection true
}

project "infra" {
  path "/Users/example/work/infra"
  layout "compact"
}

project "ops" {
  path "/Users/example/work/ops"
  layout_file "/Users/example/.config/zellij/layouts/ops.kdl"
}
```

Supported fields:

- `defaults.shell`
- `defaults.restart_on_resurrection`
- `project.path`
- `project.session`
- `project.startup`
- `project.layout`
- `project.layout_file`
- `project.restart_on_resurrection`

`project.path` is required. If `project.session` is omitted, the session name defaults to the project name.

`project.path` and `project.layout_file` support `~` and `~/...` expansion. Other relative paths are left unchanged and passed through to `zellij`; use absolute paths or `~/...` when you want config-independent behavior.

## Layout And Startup Rules

- `layout_file` takes precedence over `layout`
- `layout` takes precedence over `startup`
- `startup` only applies when no layout is configured

If you want startup commands inside a custom zellij layout, define them in the zellij layout itself.

## Resurrection Behavior

If the target session is in `resurrection` state, `zjsh` checks `restart_on_resurrection`.

- when `restart_on_resurrection` is `true`, `zjsh connect` clears the resurrected session and recreates it from the project definition
- otherwise, `zjsh connect` attaches directly to the resurrected session

Project-level `restart_on_resurrection` overrides the default value.

This is useful for tools like Neovim that rely on swap files or other runtime state. Recreating the session can avoid errors from attaching to resurrected panes whose swap files or related state have already been cleaned up.

## Zellij Integration

Example `tmux` mode keybinding for `~/.config/zellij/config.kdl`:

```kdl
keybinds {
  tmux {
    bind "K" {
      Run "sh" "-lc" "choice=$(zjsh list | gum filter --placeholder 'zjsh' --prompt='zjsh> '); [ -n \"$choice\" ] && exec zjsh connect \"$choice\"" {
        close_on_exit true
      }
      SwitchToMode "Locked"
    }
  }
}
```

`zjsh connect` accepts the selected label directly, so no extra parsing is needed.

## Doctor

`zjsh doctor` checks:

- `zellij` binary
- `zoxide` binary
- config file presence and parse errors
- configured project paths
- configured layout files

Missing config is a warning. Missing binaries, invalid config, missing project paths, and missing layout files are failures.

## Development

```bash
make fmt
make test
make build
```

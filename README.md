# zjsh

https://github.com/user-attachments/assets/cf4e0979-5ad8-4e52-960b-3881faf552e2


`zjsh` is a selector-friendly launcher for `zellij` sessions and projects.

It is a `zellij`-focused alternative to [`sesh`](https://github.com/joshmedeski/sesh), and borrows heavily from `sesh`'s idea of collecting sessions and projects into one fuzzy-selectable workflow.

`zjsh` aggregates configured projects, live `zellij` sessions, resurrectable `zellij` sessions, and `zoxide` paths into one list. Pick one entry with `fzf`, `gum`, or another selector, then pass the selected label back to `zjsh connect`.

`zjsh` does not provide its own TUI. It is designed to compose with shell commands and `zellij` keybindings.

## Motivation

[`sesh`](https://github.com/joshmedeski/sesh) provides a fast selector-first workflow for jumping between terminal sessions and projects. `zjsh` applies the same idea to `zellij`.

Use `zjsh` when you want to:

- collect configured projects, active `zellij` sessions, resurrectable sessions, and `zoxide` paths in one list
- choose a target with your preferred selector instead of a built-in UI
- attach to an existing `zellij` session, switch sessions from inside `zellij`, or create a new session from project config
- keep project startup commands and zellij layouts in a small KDL config file

`zjsh` is inspired by `sesh`, but it is not a drop-in replacement for `sesh` and does not try to be config-compatible with it.

## Install

Install the CLI with Go:

```bash
go install github.com/saweima12/zjsh/cmd/zjsh@latest
```

Make sure your Go binary directory is in `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

If you use `GOBIN`, add that directory to `PATH` instead.

Verify the binary is available:

```bash
zjsh doctor
```

## Requirements

Install-time:

- Go 1.24 or newer, required only when installing with `go install`

Runtime:

- `zellij`, used to list, attach, switch, create, and delete sessions
- `zoxide`, used as an additional project path source

Optional:

- `fzf`, used as a selector in terminal workflows
- `gum`, used as a selector in terminal workflows

Install the runtime tools before using `zjsh`:

```bash
brew install zellij zoxide fzf gum
```

Or install them with your preferred package manager.

Go is only needed to install `zjsh` from source. After installation, `zjsh` only needs `zellij` and `zoxide` at runtime. `fzf` or `gum` is only needed if you use the selector examples in this README.

## Quick Start

Create the default config file:

```bash
zjsh config init
```

Edit the generated file:

```text
~/.config/zjsh/config.kdl
```

Validate dependencies and config:

```bash
zjsh doctor
```

List all available targets:

```bash
zjsh list
```

Connect to a target:

```bash
zjsh connect api
```

## Example Workflow

Use `zjsh list` to print selector labels:

```bash
zjsh list
```

Choose one with `fzf`:

```bash
choice=$(zjsh list | fzf --prompt='zjsh> ')
```

Connect to the selected target:

```bash
[ -n "$choice" ] && zjsh connect "$choice"
```

The selected value can be a raw target or a label from `zjsh list`:

```bash
zjsh connect api
zjsh connect "◆ api"
zjsh connect "● api"
zjsh connect "→ /Users/example/work/api"
```

## Setup

The default config path is:

```text
~/.config/zjsh/config.kdl
```

Generate it with:

```bash
zjsh config init
```

Use a custom path when needed:

```bash
zjsh config init --path ~/.config/zjsh/work.kdl
```

Example config:

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

- `defaults.shell`: shell used for generated startup layouts
- `defaults.restart_on_resurrection`: default resurrection handling
- `project.path`: project directory, required
- `project.session`: zellij session name, defaults to the project name
- `project.startup`: command to run in a generated one-pane layout
- `project.layout`: named zellij layout
- `project.layout_file`: path to a zellij layout file
- `project.restart_on_resurrection`: project-level override

`project.path` and `project.layout_file` support `~` and `~/...` expansion. Other relative paths are passed through to `zellij` unchanged.

## Add A Zellij Command

Add a keybinding to `~/.config/zellij/config.kdl` so `zjsh` can be launched from inside `zellij`.

Example using `fzf` in `tmux` mode:

```kdl
keybinds {
  tmux {
    bind "K" {
      Run "sh" "-lc" "choice=$(zjsh list | fzf --prompt='zjsh> '); [ -n \"$choice\" ] && exec zjsh connect \"$choice\"" {
        close_on_exit true
      }
      SwitchToMode "Locked"
    }
  }
}
```

Example using `gum`:

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

After updating the config, reload or restart `zellij`, then press the keybinding. `zjsh connect` accepts the selected label directly, so no extra parsing is required.

You can also create a shell command outside `zellij`:

```sh
zj() {
  choice=$(zjsh list | fzf --prompt='zjsh> ')
  [ -n "$choice" ] && zjsh connect "$choice"
}
```

## Commands

List selector labels:

```bash
zjsh list
```

List full metadata as JSON:

```bash
zjsh list --json
```

Connect to a project, session, or path:

```bash
zjsh connect <target>
```

Validate dependencies, config parsing, project paths, and layout files:

```bash
zjsh doctor
```

Create a sample config file. Existing files are not overwritten:

```bash
zjsh config init [--path <file>]
```

## Entry Labels

Plain `zjsh list` output is a single column designed for selectors:

- `◆ <name>`: configured project
- `● <name>`: live `zellij` session
- `↺ <name>`: resurrectable `zellij` session
- `→ <path>`: `zoxide` path

Configured projects are identified by project path and project name, so two configured projects are not merged just because they resolve to the same session basename. Other sources are merged when they refer to the same project path or session name.

When sources merge, configured project fields are preferred while live session state from `zellij` is preserved. Any existing `zellij` session, including one already merged into a configured project, is preferred over a `zoxide` path with the same basename.

Path-only `zoxide` entries do not merge with each other by basename, so different directories with the same basename stay separate.

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

Layout and startup precedence:

1. `layout_file`
2. `layout`
3. `startup`
4. project path only

If you want startup commands inside a custom zellij layout, define them in the zellij layout itself.

## Resurrection Behavior

If the target session is in `resurrection` state, `zjsh` checks `restart_on_resurrection`.

- when `restart_on_resurrection` is `true`, `zjsh connect` clears the resurrected session and recreates it from the project definition
- otherwise, `zjsh connect` attaches directly to the resurrected session

Project-level `restart_on_resurrection` overrides the default value.

This is useful for tools like Neovim that rely on swap files or other runtime state. Recreating the session can avoid errors from attaching to resurrected panes whose swap files or related state have already been cleaned up.

## Troubleshooting

- `zjsh: command not found`: make sure your Go binary directory is in `PATH`, usually `$(go env GOPATH)/bin` or your custom `GOBIN`.
- `zjsh doctor` reports missing `zellij`: install `zellij` and make sure it is available in `PATH`.
- `zjsh doctor` reports missing `zoxide`: install `zoxide` and make sure it is available in `PATH`.
- `zjsh list` does not show a configured project: check `~/.config/zjsh/config.kdl`, confirm `project.path` exists, then run `zjsh doctor`.
- The zellij keybinding does not work: `Run "sh" "-lc"` may use a different `PATH` than your interactive shell, so use an absolute path to `zjsh` or ensure the binary directory is available to non-interactive shells.
- `startup` is ignored when a layout is configured: `layout_file` and `layout` take precedence over `startup`.

## Release

For maintainers, publish a new version by tagging a commit:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Users can install a specific version with:

```bash
go install github.com/saweima12/zjsh/cmd/zjsh@v0.1.0
```

Users can install the latest tagged version with:

```bash
go install github.com/saweima12/zjsh/cmd/zjsh@latest
```

## Development

```bash
make fmt
make test
make build
```

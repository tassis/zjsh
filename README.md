# zjsh

[https://github.com/user-attachments/assets/cf4e0979-5ad8-4e52-960b-3881faf552e2](https://github.com/user-attachments/assets/cf4e0979-5ad8-4e52-960b-3881faf552e2)

A sesh-like session launcher for `zellij`.

`zjsh` collects configured projects, live `zellij` sessions, resurrectable sessions, optional `zoxide` paths, and the current directory into one selector-friendly list.

It can also list configured macros and execute them in a new pane inside the current `zellij` session.

It does not provide its own TUI. Instead, it is designed to compose with `fzf`, `gum`, shell scripts, and `zellij` keybindings.

```sh
choice=$(zjsh list -i | fzf --prompt='zjsh> ')
[ -n "$choice" ] && zjsh connect "$choice"
```

## Features

* List configured projects from the platform default config path
* List live and resurrectable `zellij` sessions
* Optionally list recent paths from `zoxide`
* Always provide `.` as a current-directory target
* Define reusable current-directory workflows with `cwd true`
* Use `layout` or `layout_file` per project
* Define reusable macros with `cwd` and `exec`
* Compose with shell scripts, selectors, and `zellij` keybindings

## Install

```sh
go install github.com/tassis/zjsh/cmd/zjsh@latest
```

Make sure your Go binary directory is in `PATH`:

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

Go is only required when installing with `go install`; the built binary does not require Go at runtime.

## Requirements

Required at runtime:

* `zellij`

Optional:

* `zoxide`, used as an additional path source
* `fzf` or `gum`, used by the examples below

Example install with Homebrew:

```sh
brew install zellij zoxide fzf gum
```

The shell examples in this README use POSIX shell syntax. On Windows, use PowerShell equivalents or run them in a Unix-like shell.

## Quick Start

Create a config file:

```sh
zjsh config init
```

Check your setup:

```sh
zjsh doctor
```

List targets:

```sh
zjsh list
zjsh list -i
```

Connect to a project, session, path, or the current directory:

```sh
zjsh connect api
zjsh connect .
```

List and execute macros from inside `zellij`:

```sh
zjsh macros
choice=$(zjsh macros -i | fzf --prompt='macro> ')
[ -n "$choice" ] && zjsh exec "$choice"
```

## Config

`zjsh config init` writes an OS-appropriate sample config.

Default config path:

```text
Linux/macOS: ~/.config/zjsh/config.kdl
Windows: %AppData%\zjsh\config.kdl
```

Example:

```kdl
defaults {
  shell "sh"
  restart_on_resurrection false
}

project "api" {
  path "~/work/api"
  session "api"
}

project "infra" {
  path "~/work/infra"
  layout "compact"
}

project "ops" {
  path "~/work/ops"
  layout_file "~/.config/zellij/layouts/ops.kdl"
}

project "scratch" {
  cwd true
  session "scratch"
  layout "compact"
}

macro "prod" {
  exec "ssh" "prod"
}

macro "api-shell" {
  cwd "~/work/api"
  exec "docker" "compose" "exec" "api" "sh"
}
```

Each project must use exactly one path mode:

```kdl
project "static" {
  path "~/work/static"
}

project "dynamic" {
  cwd true
}
```

Supported project fields:

* `path`: static project directory
* `cwd`: use the runtime current directory when set to `true`
* `session`: zellij session name, defaults to the project name
* `layout`: named zellij layout
* `layout_file`: path to a zellij layout file
* `restart_on_resurrection`: project-level resurrection behavior

Supported defaults:

* `shell`
* `restart_on_resurrection`
* `icon_project`
* `icon_session`
* `icon_resurrectable`
* `icon_path`
* `icon_macro`

Supported macro fields:

* `cwd`: run the macro in this directory; defaults to the runtime current directory
* `exec`: command argv to execute in a new pane

`path`, `layout_file`, and macro `cwd` support `~` expansion.

For `exec`, the recommended form is explicit argv:

```kdl
macro "prod" {
  exec "ssh" "prod"
}
```

Simple single-string commands are also supported:

```kdl
macro "prod" {
  exec "ssh prod"
}
```

Single-string `exec` is intentionally strict. It only supports simple whitespace splitting. If you need quotes or exact argv control, use the multi-argument form.

## Current Directory

`zjsh` always includes `.` as a built-in current-directory target:

```sh
zjsh connect .
```

Use this when you are already inside a directory and want to open it in `zellij` without adding config first.

Use `cwd true` when you want a reusable current-directory template:

```kdl
project "scratch" {
  cwd true
  session "scratch"
  layout "compact"
}
```

Difference:

| Target             | Meaning                                                   |
| ------------------ | --------------------------------------------------------- |
| `.`                | open the current directory directly                       |
| `cwd true` project | open the current directory using a named project template |

For `cwd true` projects, the session name is `project.session`, or the project name if `session` is not set. It does not default to the current directory basename.

## Zellij Keybinding

Example using `fzf`:

```kdl
keybinds {
  tmux {
    bind "K" {
      Run "sh" "-lc" "choice=$(zjsh list -i | fzf --prompt='zjsh> '); [ -n \"$choice\" ] && exec zjsh connect \"$choice\"" {
        name "zjsh"
        floating true
        close_on_exit true
      }
      SwitchToMode "Locked"
    }
  }
}
```

You can use `gum filter` instead of `fzf`:

```sh
choice=$(zjsh list -i | gum filter --placeholder 'zjsh' --prompt='zjsh> ')
[ -n "$choice" ] && zjsh connect "$choice"
```

Example shell helper using `fzf`:

```sh
zj() {
  choice=$(zjsh list -i | fzf --prompt='zjsh> ')
  [ -n "$choice" ] && zjsh connect "$choice"
}
```

Example shell helper using `gum`:

```sh
zjg() {
  choice=$(zjsh list -i | gum filter --placeholder 'zjsh' --prompt='zjsh> ')
  [ -n "$choice" ] && zjsh connect "$choice"
}
```

Example macro helper using `fzf` inside `zellij`:

```sh
zjm() {
  choice=$(zjsh macros -i | fzf --prompt='macro> ')
  [ -n "$choice" ] && zjsh exec "$choice"
}
```

## Commands

```sh
zjsh list          # print raw target names/paths for selectors
zjsh list -i       # print display labels with icons
zjsh list --json   # print target metadata as JSON
zjsh macros        # print configured macro names for selectors
zjsh macros -i     # print macro labels with icons
zjsh macros --json # print macro metadata as JSON
zjsh connect NAME  # connect to a target
zjsh connect .     # connect to the current directory
zjsh exec NAME     # run a macro in a new zellij pane
zjsh doctor        # check dependencies and config
zjsh config init   # create sample config
```

## Entry Labels

`zjsh list -i` uses icons to show where each entry came from:

```text
● api
◆ infra
↺ old-api
▶ prod
→ .
→ /Users/example/work/tooling
```

Default labels:

* `●`: live `zellij` session
* `◆`: configured project
* `↺`: resurrectable session
* `▶`: configured macro
* `→`: path entry, including `.` and zoxide paths

Display order:

1. live sessions
2. configured projects
3. resurrectable sessions
4. current directory `.`
5. zoxide paths

`.` always remains a separate visible entry. It does not merge with matching zoxide paths or configured project paths.

When a configured project and a session use the same session name, they are shown as one entry.

* live session + project: the merged entry is shown primarily as the live session
* resurrectable session + project: the merged entry remains project-first

## Connect Behavior

`zjsh connect` accepts both raw values and icon labels:

```sh
zjsh connect api
zjsh connect "◆ api"
zjsh connect "● api"
zjsh connect "→ /Users/example/work/api"
zjsh connect "→ ."
```

Resolution order:

1. project name
2. session name
3. full path
4. `.` current-directory target

Layout precedence:

1. `layout_file`
2. `layout`
3. no layout; use the target directory only

For path-based entries such as zoxide paths and `.`, the session name is based on the path basename. If that name is already reserved, `zjsh` appends a short path hash.

## Macro Behavior

`zjsh macros` is separate from `zjsh list`.

Use `zjsh list` for connectable targets and `zjsh macros` for executable macro targets.

`zjsh exec` accepts both raw values and icon labels:

```sh
zjsh exec prod
zjsh exec "▶ prod"
```

Execution rules:

1. `zjsh exec` must be run inside `zellij`
2. the macro opens a new pane with `zellij action new-pane`
3. macro `cwd` is used when configured
4. otherwise the runtime current directory is used

`zjsh exec` does not create or switch sessions. It runs only in the currently attached `zellij` session.

## Resurrection

If a selected target already has a resurrectable `zellij` session, `zjsh` checks `restart_on_resurrection`.

This option only applies to resurrectable sessions. It does not affect live sessions or normal project/path launches.

* `true`: delete the resurrected session and recreate it from project config
* `false`: attach to the resurrected session directly

Project-level settings override `defaults.restart_on_resurrection`.

## Troubleshooting

* `zjsh: command not found`: make sure `$(go env GOPATH)/bin` or `GOBIN` is in `PATH`.
* Missing `zellij`: install `zellij`; this is required.
* Missing `zoxide`: this is only a warning; zoxide paths will be unavailable.
* `zjsh list` only shows `.`: no config, no sessions, and no zoxide paths were found.
* `cwd true` uses the wrong session name: it uses `session`, then project name, not the current directory basename.
* `zjsh exec` says it must run inside `zellij`: attach to or start a `zellij` session first.
* Macro `exec` parsing fails for quoted single-string commands: use the multi-argument `exec "cmd" "arg1" "arg2"` form.

## Development

```sh
make fmt
make test
make build
```

## Project status

`zjsh` is currently feature-complete for my intended workflow.

The project is expected to stay small and focused. Future changes will mainly be limited to:

- compatibility updates for new Zellij releases
- bug fixes
- documentation improvements
- small UX improvements that fit the current design
- new features only when Zellij exposes capabilities that make them practical

`zjsh` is not intended to become a full TUI session manager or a replacement for Zellij's built-in session manager.

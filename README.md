# Transparent Pipe (tp)
[![Go Report Card](https://goreportcard.com/badge/github.com/minefuto/tp)](https://goreportcard.com/report/github.com/minefuto/tp)
[![build](https://github.com/minefuto/tp/actions/workflows/build.yml/badge.svg)](https://github.com/minefuto/tp/actions/workflows/build.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/minefuto/tp)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/minefuto/tp)
![GitHub](https://img.shields.io/github/license/minefuto/tp)

This project is inspired by [akavel/up](https://github.com/akavel/up).
`tp` is a terminal-based application that displays the result of commands at every keystroke.
It makes it easy to chain commands for tasks like text manipulation.

It provides two displays:
1. The stdout of the last pipe.
2. A live preview of the current command's result.

<img src="https://github.com/minefuto/tp/blob/main/gif/tp.gif">

## Installation
```sh
go install github.com/minefuto/tp@latest
```

## Keybindings
| Operation                                 | Key                                      |
|-------------------------------------------|------------------------------------------|
| Move left (one char)                      | <kbd>←</kbd> / <kbd>Ctrl-B</kbd>         |
| Move right (one char)                     | <kbd>→</kbd> / <kbd>Ctrl-F</kbd>         |
| Move left (one word)                      | <kbd>Alt←</kbd> / <kbd>Alt-B</kbd>       |
| Move right (one word)                     | <kbd>Alt→</kbd> / <kbd>Alt-F</kbd>       |
| Move to beginning of line                 | <kbd>Home</kbd> / <kbd>Ctrl-A</kbd>      |
| Move to end of line                       | <kbd>End</kbd> / <kbd>Ctrl-E</kbd>       |
| Delete one char before the cursor         | <kbd>Backspace</kbd> / <kbd>Ctrl-H</kbd> |
| Delete one char after the cursor          | <kbd>Delete</kbd> / <kbd>Ctrl-D</kbd>    |
| Delete one word before the cursor         | <kbd>Ctrl-W</kbd>                        |
| Delete from the cursor to end of line     | <kbd>Ctrl-K</kbd>                        |
| Delete entire line                        | <kbd>Ctrl-U</kbd>                        |


## Sandbox
`tp` executes commands at every keystroke, so all preview commands run inside a sandbox that restricts file system access to read-only. This prevents destructive operations such as `rm` or any other write to the file system.

| OS    | Sandbox mechanism                                                                     | Requirement                      |
|-------|---------------------------------------------------------------------------------------|----------------------------------|
| macOS | [Apple Seatbelt](https://www.unix.com/man-page/osx/1/sandbox-exec/) (`sandbox-exec`) | `sandbox-exec` must be available |
| Linux | [Landlock](https://docs.kernel.org/userspace-api/landlock.html)                       | Kernel with Landlock v3+ support |

`tp` will exit with an error if the required sandbox is not available.

## Shell Integration
You can synchronize your shell's line buffer with `tp`'s input field.
The following config enables `zsh` integration with the keybinding `ctrl + |`:
```zsh
function transparent-pipe() {
  BUFFER="$(tp -c "${BUFFER}|")"
  CURSOR=$#BUFFER
}
zle -N transparent-pipe
bindkey "^|" transparent-pipe
```

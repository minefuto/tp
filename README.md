# Transparent Pipe(tp)
[![Go Report Card](https://goreportcard.com/badge/github.com/minefuto/tp)](https://goreportcard.com/report/github.com/minefuto/tp)
[![build](https://github.com/minefuto/tp/actions/workflows/build.yml/badge.svg)](https://github.com/minefuto/tp/actions/workflows/build.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/minefuto/tp)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/minefuto/tp)
![GitHub](https://img.shields.io/github/license/minefuto/tp)

This project is inspired by [akavel/up](https://github.com/akavel/up).  
`tp` is a terminal-based application for display the result of the commands at every keystroke.  
Make it easy to chain commands for such as string manipulation.

It provides two displays.  
1. stdout from last pipe.  
2. live preview of the command result.  

<img src="https://github.com/minefuto/tp/blob/main/gif/tp.gif">

## Shell integration
It can be synchronized shell's linebuffer and `tp`'s inputfield.  
The below config is the `zsh` integration.  
keymappings: `ctrl + |`
```
function transparent-pipe() {
  BUFFER="$(tp -c "${BUFFER}|")"
  CURSOR=$#BUFFER
}
zle -N transparent-pipe
bindkey "^|" transparent-pipe
```

## Limitation
`tp` executes the command at every keystroke. There is possibility to execute dangerous commands such as `rm`.  
So, `tp` is only supported specific string manipulation commands. Any other commands will be executed when you pressed `|`, not every keystroke.  
Also, `tp` is not supported redirections(`<`, `>`).  

supported commands:  
`awk`,`cut`,`egrep`,`grep`,`head`,`jq`,`nl`,`sed`,`sort`,`tail`,`tr`,`uniq`,`vgrep`,`wc`,`yq`  

## Installation
```
$ go install github.com/minefuto/tp@latest
```

## Keybindings
| operation                                 | key                                      |
|-------------------------------------------|------------------------------------------|
| Move left(one char)                       | <kbd>←</kbd> / <kbd>Ctrl-B</kbd>         |
| Move right(one char)                      | <kbd>→</kbd> / <kbd>Ctrl-F</kbd>         |
| Move left(one word)                       | <kbd>Alt←</kbd> / <kbd>Alt-B</kbd>       |
| Move right(one word)                      | <kbd>Alt→</kbd> / <kbd>Alt-F</kbd>       |
| Move begin of the line                    | <kbd>Home</kbd> / <kbd>Ctrl-A</kbd>      |
| Move end of the line                      | <kbd>End</kbd> / <kbd>Ctrl-E</kbd>       |
| Delete one char before the cursor         | <kbd>Backspace</kbd> / <kbd>Ctrl-H</kbd> |
| Delete one char after the cursor          | <kbd>Delete</kbd> / <kbd>Ctrl-D</kbd>    |
| Delete one word before the cursor         | <kbd>Ctrl-W</kbd>                        |
| Delete from the cursor to end of the line | <kbd>Ctrl-K</kbd>                        |
| Delete all line                           | <kbd>Ctrl-U</kbd>                        |

## Options
```
> tp -h
Usage of tp:
  -c, --command        Return commandline text
  -h, --help           Show help
      --horizontal     Split the view horizontally
  -s, --shell string   Select a shell to use (default "$SHELL")
  -v, --version        Show version
```

## Supported OS
macOS, Linux

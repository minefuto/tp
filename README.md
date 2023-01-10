# Transparent Pipe(tp)
[![Go Report Card](https://goreportcard.com/badge/github.com/minefuto/tp)](https://goreportcard.com/report/github.com/minefuto/tp)
[![release](https://github.com/minefuto/tp/actions/workflows/release.yml/badge.svg)](https://github.com/minefuto/tp/actions/workflows/release.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/minefuto/tp)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/minefuto/tp)
![GitHub](https://img.shields.io/github/license/minefuto/tp)

This project is inspired by [akavel/up](https://github.com/akavel/up).  
`tp` is a terminal-based application for display the result of the command in real-time with each keystroke.  

It provides two displays.  
1. the input passed from last pipeline.  
2. the output of the command currently being typed.  

You can consider the next commands while watching the input passed from pipeline.  
These will help you create complex commands including pipelines for get the ideal output with try and error.  
<br>
Please type <kbd>Enter</kbd> when you completed to create command in `tp`.  
Then, `tp` returns the full result of the command as stdout/stderr.  

<img src="https://github.com/minefuto/tp/blob/main/gif/tp.gif">

Also, `tp` can collaborate with the shell.  
By typing a shortcut key, you can start `tp` by capturing the command being typed into shell.  
And the command being typed into `tp` return to shell when type <kbd>Enter</kbd>.  

<img src="https://github.com/minefuto/tp/blob/main/gif/tp-shell.gif">

If you want to collaborate with the shell, please add the following to shell's config file.  
`<key>`: Specify any shortcut key.  

Bash
```
function transparent-pipe() {
  READLINE_LINE=$(tp -c "${READLINE_LINE}")
  READLINE_POINT=${#READLINE_LINE}
}
bind -x '"<key>": transparent-pipe'
```
Zsh
```
function transparent-pipe() {
  BUFFER="$(tp -c "${BUFFER}")"
  CURSOR=$#BUFFER
}
zle -N transparent-pipeline
bindkey "<key>" transparent-pipe
```
Fish
```
function transparent-pipe
  commandline | read -l buffer
  commandline -r (tp -c "$buffer")
  commandline -f repaint
end
function fish_user_key_bindings
  bind "<key>" transparent-pipe
end
```
<br>

**Warning!!!**  
`tp` executes the command being typed with each keystroke. There is possibility to execute dangerous commands.  
So, create/delete operations(such as `mkdir`, `rm`) should not be typed because you might execute a careless command.  
`tp` is not designed for such operations.  

But I'm afraid of typo.  
`tp` provides a feature of prevent execution a specific commands.  
Please create block command list in `$TP_BLOCK_COMMAND` with `:` as delimiter. For example,  
```
export TP_BLOCK_COMMAND='mkdir:rmdir:rm:mv'
```
Also, disable keystroke of redirection(`<`, `>`) in `tp` for the same reason.  

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
  -c, --command        Return commandline text (for collaborate with the shell)
  -h, --help           Show help
      --horizontal     Split the view horizontally
  -s, --shell string   Specify the shell to use (default "$SHELL")
  -v, --version        Show version
```

## Supported OS
macOS, Linux

# Transparent Pipe : A terminal-based pipeline command
[![Go Report Card](https://goreportcard.com/badge/github.com/minefuto/tp)](https://goreportcard.com/report/github.com/minefuto/tp)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/minefuto/tp/build)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/minefuto/tp)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/minefuto/tp)
![GitHub](https://img.shields.io/github/license/minefuto/tp)

This is a terminal-based application for interactively previews the stdin/stdout around the pipeline.

## Installation
```
$ git clone https://github.com/minefuto/tp.git
$ cd tp
$ make install
```

## Usage
### 1. commandline mode  
<img src="https://github.com/minefuto/tp/blob/main/gif/mode1.gif">

Add the following to shell's config file.  
`<key>`: Specify any key.  
**Bash(`.bashrc`)**
```
function transparent-pipeline() {
  READLINE_LINE=$(tp -c "${READLINE_LINE}|")
  READLINE_POINT=${#READLINE_LINE}
}
bind -x '"<key>": transparent-pipeline'
```
**Zsh(`.zshrc`)**
```
function transparent-pipeline() {
  BUFFER="$(tp -c "${BUFFER}|")"
  CURSOR=$#BUFFER
}
zle -N transparent-pipeline
bindkey "<key>" transparent-pipeline
```
**Fish(`config.fish`)**
```
function transparent-pipeline
  commandline | read -l buffer
  commandline -r (tp -c "$buffer|")
  commandline -f repaint
end
function fish_user_key_bindings
  bind "<key>" transparent-pipeline
end
```

### 2. stdin/stdout mode  
<img src="https://github.com/minefuto/tp/blob/main/gif/mode2.gif">

## Options
```
> tp -h
Usage of tp:
  -c, --command        Return commandline text (Please see the "1. commandline mode")
  -h, --help           Show help
      --horizontal     Split the view horizontally
  -s, --shell string   Specify the shell to use (default "$SHELL")
  -v, --version        Show version
```

## Supported OS
macOS, Linux

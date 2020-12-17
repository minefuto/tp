# tp - Transparent Pipeline
![GitHub](https://img.shields.io/github/license/minefuto/tp?style=for-the-badge)

This is a tool that interactively previews the command's stdin/stdout.

## Installation
```
$ go get github.com/minefuto/tp 
```

## Usage
### 1. stdin/stdout mode  
<img src="https://github.com/minefuto/tp/blob/main/gif/mode1.gif">

### 2. commandline mode  
<img src="https://github.com/minefuto/tp/blob/main/gif/mode2.gif">

Add the following to shell's config file.  
**Zsh(`.zshrc`)**
```
function transparent-pipeline() {
  BUFFER="$(tp -c $BUFFER)"
  CURSOR=$#BUFFER
}
zle -N transparent-pipeline
bindkey '^t' transparent-pipeline
```

**Fish(`config.fish`)**
```
function transparent-pipeline
  commandline | read -l buffer
  commandline (tp -c "$buffer")
end
function fish_user_key_bindings
  bind \ct transparent-pipeline
end
```

## Supported OS
macOS, Linux

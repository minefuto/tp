package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/text/transform"
)

const (
	name    = "tp"
	version = "0.1.0"
)

var (
	shell       string
	initCommand string
	commandFlag bool
	helpFlag    bool
	versionFlag bool
	stdinBytes  = []byte("")
)

type tui struct {
	*tview.Application
	cliPane    *cliPane
	stdinPane  *viewPane
	stdoutPane *viewPane
}

func newTui() *tui {
	cliPane := newCliPane()
	stdinPane := newViewPane("stdin")
	stdoutPane := newViewPane("stdout")

	viewPanes := tview.NewFlex()
	viewPanes.SetDirection(tview.FlexColumn).
		AddItem(stdinPane, 0, 1, false).
		AddItem(stdoutPane, 0, 1, false)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).
		AddItem(cliPane, 1, 1, false).
		AddItem(viewPanes, 0, 1, false)

	t := &tui{
		Application: tview.NewApplication(),
		cliPane:     cliPane,
		stdinPane:   stdinPane,
		stdoutPane:  stdoutPane,
	}
	t.SetRoot(flex, true).SetFocus(cliPane)
	return t
}

func (t *tui) start() int {
	t.setAction()

	go func() {
		if t.cliPane.prompt == "" {
			t.stdinPane.setData(stdinBytes)
		} else {
			t.stdinPane.execCommand(t.cliPane.prompt, stdinBytes)
		}
		t.stdoutPane.execCommand(t.cliPane.GetText(), t.stdinPane.data)
	}()

	if err := t.Run(); err != nil {
		t.Stop()
		return 1
	}
	return 0
}

func (t *tui) setAction() {
	t.cliPane.SetChangedFunc(func(text string) {
		if !t.cliPane.skip && !t.cliPane.disable {
			go func() {
				t.stdoutPane.execCommand(t.cliPane.GetText(), t.stdinPane.data)
			}()
		}
		if t.cliPane.skip {
			t.cliPane.skip = false
		}
	})

	t.stdinPane.SetChangedFunc(func() {
		t.Draw()
	})

	t.stdoutPane.SetChangedFunc(func() {
		t.Draw()
	})

	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			if commandFlag {
				fmt.Println(initCommand)
				return event
			}
			return event

		case tcell.KeyEnter:
			t.Stop()
			if commandFlag {
				fmt.Println(adjustPipe(t.cliPane.prompt) + t.cliPane.GetText())
				return nil
			}
			t.stdoutPane.syncUpdate(func() {
				fmt.Print(string(t.stdoutPane.data))
			})
			return nil

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if t.cliPane.GetText() == "" {
				if t.cliPane.prompt == "" {
					return event
				}
				go func() {
					t.stdoutPane.ctx, t.stdoutPane.cancel = nil, nil

					t.cliPane.disableHandler(func() {
						t.cliPane.setPrompt(t.cliPane.prompt)
					})
					t.stdinPane.syncUpdate(func() {
						t.stdoutPane.setData(t.stdinPane.data)
					})

					if t.cliPane.prompt == "" {
						t.stdinPane.setData(stdinBytes)
					} else {
						t.stdinPane.execCommand(t.cliPane.prompt, stdinBytes)
					}
				}()
				return nil
			}
			return event

		case tcell.KeyRune:
			switch event.Rune() {
			case '|':
				go func() {
					t.cliPane.disableHandler(func() {
						t.cliPane.addPrompt()
						t.stdoutPane.Clear()
						t.stdoutPane.syncUpdate(func() {
							t.stdinPane.setData(t.stdoutPane.data)
						})
					})
					t.stdoutPane.execCommand(t.cliPane.GetText(), t.stdinPane.data)
				}()
				return nil
			case ' ':
				t.cliPane.skipHandler()
			}
		}
		return event
	})
}

type cliPane struct {
	*tview.InputField
	symbol  string
	prompt  string
	disable bool
	skip    bool
}

func newCliPane() *cliPane {
	inputField := tview.NewInputField()
	inputField.SetAcceptanceFunc(tview.InputFieldMaxLength(200)).
		SetFieldWidth(0)

	symbol := "| "
	if bytes.Equal(stdinBytes, []byte("")) {
		symbol = "> "
	}

	c := &cliPane{
		InputField: inputField,
		symbol:     symbol,
		disable:    false,
		skip:       false,
	}
	c.setPrompt(initCommand)

	return c
}

func (c *cliPane) skipHandler() {
	c.skip = true
}

func (c *cliPane) disableHandler(fn func()) {
	c.disable = true
	fn()
	c.disable = false
}

func (c *cliPane) setPrompt(text string) {
	if strings.Contains(text, "|") {
		c.prompt = text[:strings.LastIndex(text, "|")]
		c.SetLabel(c.symbol + adjustPipe(c.prompt))
		c.SetText(text[strings.LastIndex(text, "|")+1:])
		return
	}
	c.SetLabel(c.symbol)
	c.SetText(text)
	c.prompt = ""
}

func (c *cliPane) addPrompt() {
	c.prompt = adjustPipe(c.prompt) + c.GetText()
	c.SetLabel(c.symbol + adjustPipe(c.prompt)).
		SetText("")
}

func adjustPipe(text string) string {
	if text == "" {
		return ""
	}
	return text + "|"
}

type viewPane struct {
	*tview.TextView
	data   []byte
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

func newViewPane(name string) *viewPane {
	textView := tview.NewTextView()
	textView.SetWrap(false).
		SetScrollable(false).
		SetTitleAlign(tview.AlignLeft).
		SetTitle(name).
		SetBorder(true)

	v := &viewPane{
		TextView: textView,
		data:     []byte(""),
	}
	return v
}

func (v *viewPane) setData(inputBytes []byte) {
	v.reset()

	tt := newTextLineTransformer()
	w := transform.NewWriter(v, tt)

	v.syncUpdate(func() {
		v.data = make([]byte, len(inputBytes))
		copy(v.data, inputBytes)
		io.Copy(w, bytes.NewReader(inputBytes))
	})
}

func (v *viewPane) execCommand(text string, inputBytes []byte) {
	v.reset()

	_data := new(bytes.Buffer)
	tt := newTextLineTransformer()
	w := transform.NewWriter(v, tt)
	mw := io.MultiWriter(w, _data)

	cmd := exec.CommandContext(v.ctx, shell, "-c", text)
	cmd.Stdin = bytes.NewReader(inputBytes)
	cmd.Stdout = mw

	v.syncUpdate(func() {
		cmd.Run()
		v.data = _data.Bytes()
	})
}

func (v *viewPane) syncUpdate(fn func()) {
	v.mu.Lock()
	fn()
	v.mu.Unlock()
}

func (v *viewPane) reset() {
	v.Clear()

	if v.cancel != nil {
		v.cancel()
	}
	v.ctx, v.cancel = context.WithCancel(context.Background())
}

type textLineTransformer struct {
	transform.NopResetter
	line  int
	limit int
	temp  []byte
}

func newTextLineTransformer() *textLineTransformer {
	_, height, _ := terminal.GetSize(int(os.Stderr.Fd()))
	tt := &textLineTransformer{
		line:  0,
		limit: height - 3,
		temp:  []byte(""),
	}
	return tt
}

func (tt *textLineTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if tt.limit <= tt.line {
		nSrc = len(src)
		nDst = 0
		return
	}

	nSrc = len(src)
	_src := src
	if len(tt.temp) > 0 {
		_src = make([]byte, len(tt.temp)+len(src))
		copy(_src, tt.temp)
		copy(_src[len(tt.temp):], src)
	}

	if len(_src) > 4096 {
		tt.temp = make([]byte, len(_src[4096:]))
		copy(tt.temp, _src[4096:])
		err = transform.ErrShortDst
	}

	i, b := 0, 0
	for {
		i = bytes.IndexByte(_src[b:], '\n')
		if i == -1 {
			nDst = copy(dst, _src)
			return
		}
		b = b + i + 1

		if b >= 4096 {
			nDst = copy(dst, _src)
			return
		}
		tt.line++

		if tt.limit <= tt.line {
			nDst = copy(dst, _src[:b-1])
			return
		}
	}
}

func main() {
	runewidth.DefaultCondition.EastAsianWidth = false

	flag.BoolVarP(&helpFlag, "help", "h", false, "Show help")
	flag.BoolVarP(&versionFlag, "version", "v", false, "Show version")
	flag.BoolVarP(&commandFlag, "command", "c", false, "Return commandline text")
	flag.StringVarP(&shell, "shell", "s", os.Getenv("SHELL"), "Specify the shell to use")
	flag.Parse()

	if helpFlag {
		fmt.Println("Usage of tp:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if versionFlag {
		fmt.Printf("%s version %s\n", name, version)
		os.Exit(0)
	}

	if shell == "" {
		log.Fatalln("Error: Shell is not found, please specify the shell by \"-s\" option")
	}

	initCommand = flag.Arg(0)

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		stdinBytes, _ = ioutil.ReadAll(os.Stdin)
	}

	t := newTui()
	os.Exit(t.start())
}

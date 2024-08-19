package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	flag "github.com/cornfeedhobo/pflag"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/text/transform"
)

var (
	name    = "tp"
	version = ""
)

var (
	shell         string
	initCommand   string
	commandFlag   bool
	helpFlag      bool
	versionFlag   bool
	stdinBytes    []byte
	allowCommands = [...]string{"awk", "cut", "egrep", "grep", "head", "jq", "nl", "sed", "sort", "tail", "tr", "uniq", "vgrep", "wc", "yq"}
)

var getTerminalHeight = func() int {
	_, height, _ := terminal.GetSize(int(os.Stderr.Fd()))
	return height
}

type tui struct {
	*tview.Application
	cliPane    *cliPane
	stdinPane  *stdinViewPane
	stdoutPane *stdoutViewPane
}

func newTui() *tui {
	cliPane := newCliPane()
	stdinPane := newStdinViewPane()
	stdoutPane := newStdoutViewPane()

	flex := tview.NewFlex()
	viewPanes := tview.NewFlex()
	viewPanes.SetDirection(tview.FlexColumn).
		AddItem(stdinPane, 0, 1, false).
		AddItem(stdoutPane, 0, 1, false)

	flex.SetDirection(tview.FlexRow).
		AddItem(cliPane, 1, 0, false).
		AddItem(viewPanes, 0, 1, false)

	t := &tui{
		Application: tview.NewApplication(),
		cliPane:     cliPane,
		stdinPane:   stdinPane,
		stdoutPane:  stdoutPane,
	}
	t.SetRoot(flex, true).SetFocus(cliPane)
	t.setAction()
	return t
}

func (t *tui) setAction() {
	t.stdinPane.SetChangedFunc(func() {
		t.Draw()
	})

	t.stdoutPane.SetChangedFunc(func() {
		t.Draw()
	})

	t.cliPane.SetChangedFunc(func(text string) {
		_text := strings.TrimSpace(text)
		if t.cliPane.trimText == _text {
			t.cliPane.trimText = _text
			return
		}
		t.cliPane.trimText = _text
		t.stdoutPane.reset()
		t.updateStdoutView(text)
	})

	t.cliPane.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			if commandFlag {
				fmt.Println(initCommand)
				return event
			}
			return event

		case tcell.KeyEnter:
			t.stdinPane.cancel()
			t.stdoutPane.cancel()
			t.Stop()

			_text := adjustPipe(t.cliPane.prompt) + t.cliPane.GetText()
			if commandFlag {
				fmt.Println(_text)
				return nil
			}

			cmd := exec.Command(shell, "-c", _text)
			cmd.Stdin = bytes.NewReader(stdinBytes)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			return nil

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if t.cliPane.GetText() == "" {
				if t.cliPane.prompt == "" {
					return event
				}
				t.cliPane.setPrompt(t.cliPane.prompt)
				t.stdinPane.reset()
				t.updateStdinView()
				return nil
			}
			return event

		case tcell.KeyRune:
			switch event.Rune() {
			case '|':
				t.cliPane.addPrompt()
				t.stdinPane.reset()
				t.updateStdinView()
				return nil
			case '>':
				return nil
			case '<':
				return nil
			}
		}
		return event
	})
}

func (t *tui) start() int {
	t.updateStdinView()
	t.updateStdoutView(t.cliPane.GetText())

	if err := t.Run(); err != nil {
		t.Stop()
		return 1
	}
	return 0
}

func (t *tui) updateStdinView() {
	stdinCtx, stdinCancel := context.WithCancel(t.stdinPane.ctx)

	p := t.cliPane.prompt
	go func() {
		defer stdinCancel()
		if p == "" {
			t.stdinPane.setData(stdinBytes)
		} else {
			t.stdinPane.execCommand(stdinCtx, p, stdinBytes)
		}
	}()
	go func() {
		s := spinner()
		t.stdinPane.syncUpdate(func() {
			t.stdinPane.isLoading = true
		})
		for {
			select {
			case <-stdinCtx.Done():
				t.stdinPane.syncUpdate(func() {
					t.stdinPane.isLoading = false
				})
				t.QueueUpdateDraw(func() {
					t.stdinPane.SetTitle(t.stdinPane.name)
				})
				return
			case <-time.After(100 * time.Millisecond):
				t.QueueUpdateDraw(func() {
					t.stdinPane.SetTitle(t.stdinPane.name + s())
				})
			}
		}
	}()
}

func (t *tui) updateStdoutView(text string) {
	stdoutCtx, stdoutCancel := context.WithCancel(t.stdoutPane.ctx)

	go func() {
		defer stdoutCancel()
		t.stdinPane.syncUpdate(func() {
			t.QueueUpdateDraw(func() {
				if isBlock(text) || t.stdinPane.isLoading {
					t.stdoutPane.SetTitle("no preview")
				} else {
					t.stdoutPane.SetTitle("stdout/stderr")
				}
			})
			if !isBlock(text) && !t.stdinPane.isLoading {
				t.stdoutPane.execCommand(stdoutCtx, text, t.stdinPane.data)
			}
		})
	}()
}

func spinner() func() string {
	c := 0
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return func() string {
		i := c % len(spinners)
		c++
		return spinners[i]
	}
}

type cliPane struct {
	*tview.InputField
	symbol   string
	prompt   string
	trimText string
	mu       sync.Mutex
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
	}
	c.setPrompt(initCommand)
	return c
}

func (c *cliPane) syncUpdate(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn()
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
	name   string
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

func newViewPane(name string) *viewPane {
	textView := tview.NewTextView()
	textView.SetWrap(false).
		SetDynamicColors(true).
		SetScrollable(false).
		SetTitleAlign(tview.AlignLeft).
		SetTitle(name).
		SetBorder(true)

	ctx, cancel := context.WithCancel(context.Background())

	v := &viewPane{
		TextView: textView,
		name:     name,
		ctx:      ctx,
		cancel:   cancel,
	}
	return v
}

func (v *viewPane) syncUpdate(fn func()) {
	v.mu.Lock()
	defer v.mu.Unlock()
	fn()
}

func (v *viewPane) reset() {
	v.Clear()
	v.cancel()
	v.ctx, v.cancel = context.WithCancel(context.Background())
}

type stdinViewPane struct {
	*viewPane
	data      []byte
	isLoading bool
}

func newStdinViewPane() *stdinViewPane {
	v := newViewPane("stdin")
	si := &stdinViewPane{
		viewPane:  v,
		data:      []byte(""),
		isLoading: false,
	}
	return si
}

func (si *stdinViewPane) setData(inputBytes []byte) {
	tt := newTextLineTransformer()
	w := transform.NewWriter(tview.ANSIWriter(si), tt)

	si.syncUpdate(func() {
		si.data = make([]byte, len(inputBytes))
		copy(si.data, inputBytes)
	})
	io.Copy(w, bytes.NewReader(inputBytes))
}

func (si *stdinViewPane) execCommand(ctx context.Context, text string, inputBytes []byte) {
	_data := new(bytes.Buffer)
	tt := newTextLineTransformer()
	w := transform.NewWriter(tview.ANSIWriter(si), tt)
	mw := io.MultiWriter(w, _data)

	si.syncUpdate(func() {
		si.data = []byte("")
	})
	cmd := exec.CommandContext(ctx, shell, "-c", text)

	cmd.Stdin = bytes.NewReader(inputBytes)
	cmd.Stdout = mw
	cmd.Run()

	select {
	case <-ctx.Done():
	default:
		si.syncUpdate(func() {
			si.data = _data.Bytes()
		})
	}
}

type stdoutViewPane struct {
	*viewPane
}

func newStdoutViewPane() *stdoutViewPane {
	v := newViewPane("stdout/stderr")
	so := &stdoutViewPane{
		viewPane: v,
	}
	return so
}

func (so *stdoutViewPane) execCommand(ctx context.Context, text string, inputBytes []byte) {
	tt := newTextLineTransformer()
	w := transform.NewWriter(tview.ANSIWriter(so), tt)

	cmd := exec.CommandContext(ctx, shell, "-c", text)

	cmd.Stdin = bytes.NewReader(inputBytes)
	cmd.Stdout = w
	cmd.Stderr = w

	cmd.Run()
}

func isBlock(text string) bool {
	for _, cmd := range allowCommands {
		_text := strings.TrimLeft(text, " ")
		if _text == cmd {
			return false
		}
		if strings.HasPrefix(_text, cmd+" ") {
			return false
		}
	}
	return true
}

type textLineTransformer struct {
	transform.NopResetter
	line  int
	limit int
	temp  []byte
}

func newTextLineTransformer() *textLineTransformer {
	tt := &textLineTransformer{
		line:  0,
		limit: getTerminalHeight() - 3,
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
	flag.BoolVarP(&helpFlag, "help", "h", false, "Show help")
	flag.BoolVarP(&versionFlag, "version", "v", false, "Show version")
	flag.BoolVarP(&commandFlag, "command", "c", false, "Return commandline text")
	flag.StringVarP(&shell, "shell", "s", os.Getenv("SHELL"), "Select a shell to use")
	flag.Parse()

	if helpFlag {
		fmt.Fprintln(os.Stderr, "Usage of tp:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if versionFlag {
		if version == "" {
			info, ok := debug.ReadBuildInfo()
			if !ok {
				version = "(devel)"
			} else {
				version = info.Main.Version
			}
		}
		fmt.Printf("%s version %s\n", name, version)
		os.Exit(0)
	}

	if os.Getenv("SHELL") == "" {
		fmt.Fprint(os.Stderr, "$SHELL not found, please select a shell by '-s' option")
		os.Exit(1)
	}

	_, err := exec.LookPath(shell)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s not found", shell)
		os.Exit(1)
	}

	initCommand = flag.Arg(0)

	if !isatty.IsTerminal(os.Stdin.Fd()) {
		stdinBytes, _ = ioutil.ReadAll(os.Stdin)
	}

	t := newTui()
	os.Exit(t.start())
}

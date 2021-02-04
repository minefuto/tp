package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/transform"
)

func TestGetBlockList(t *testing.T) {
	var input string
	env = func() string {
		return input
	}

	cases := []struct {
		input  string
		result []string
	}{
		{input: "", result: nil},
		{input: "a", result: []string{"a"}},
		{input: "a:b", result: []string{"a", "b"}},
	}

	for _, tc := range cases {
		input = tc.input
		result := getBlockList()
		if !(reflect.DeepEqual(result, tc.result)) {
			r := `result:   "%s"`
			e := `expected: "%s"`
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestIsBlock(t *testing.T) {
	cases := []struct {
		input  string
		result bool
	}{
		{input: "rm", result: true},
		{input: "rm aaa", result: true},
		{input: "rma aaa", result: false},
	}

	for _, tc := range cases {
		if isBlock(tc.input) != tc.result {
			r := `result:   "%s"`
			e := `expected: "%s"`
			t.Errorf("\n%s\n%s", r, e)
		}
	}

}

func TestSpinner(t *testing.T) {
	cases := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	s := spinner()
	for i := 0; i < len(cases)+1; i++ {
		result := s()
		if !(result == cases[i%len(cases)]) {
			t.Errorf("result: %s, expected: %s", result, cases[i%len(cases)])
		}
	}
}

func TestSetPrompt(t *testing.T) {
	cases := []struct {
		input  string
		prompt string
		text   string
	}{
		{input: "ls", prompt: "", text: "ls"},
		{input: "ls | grep a", prompt: "ls ", text: " grep a"},
		{input: "ls | grep a | wc", prompt: "ls | grep a ", text: " wc"},
	}
	for _, tc := range cases {
		c := newCliPane()
		c.setPrompt(tc.input)
		if !(c.prompt == tc.prompt) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, c.prompt), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.prompt), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
		if !(c.GetText() == tc.text) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, c.GetText()), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.text), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestAddPrompt(t *testing.T) {
	cases := []struct {
		prompt string
		text   string
		result string
	}{
		{prompt: "", text: "ls ", result: "ls "},
		{prompt: "ls |", text: " grep a", result: "ls | grep a"},
	}
	for _, tc := range cases {
		c := newCliPane()
		c.setPrompt(tc.prompt)
		c.SetText(tc.text)
		c.addPrompt()
		if !(c.prompt == tc.result) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, c.prompt), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestSetData(t *testing.T) {
	getTerminalHeight = func() int {
		return 6
	}

	cases := []struct {
		input   string
		result  string
		result2 string
	}{
		{input: "a\n", result: "a\n", result2: "a\n"},
		{input: "a\na\na\na\na", result: "a\na\na\na\na", result2: "a\na\na"},
	}
	for _, tc := range cases {
		si := newStdinViewPane()
		si.setData([]byte(tc.input))
		if !(bytes.Equal(si.data, []byte(tc.result))) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, string(si.data)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
		if !(si.GetText(true) == tc.result2) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, si.GetText(true)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result2), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestExecCommandStdin(t *testing.T) {
	shell = "sh"
	blockCommands = nil
	getTerminalHeight = func() int {
		return 6
	}

	cases := []struct {
		cmd     string
		stdin   string
		result  string
		result2 string
	}{
		{cmd: "echo a", stdin: "", result: "a\n", result2: "a\n"},
		{cmd: "grep a", stdin: "a\nb\na\na\n", result: "a\na\na\n", result2: "a\na\na"},
	}
	for _, tc := range cases {
		si := newStdinViewPane()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		si.execCommand(ctx, tc.cmd, []byte(tc.stdin))
		if !(bytes.Equal(si.data, []byte(tc.result))) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, string(si.data)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
		if !(si.GetText(true) == tc.result2) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, si.GetText(true)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result2), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestExecCommandStdout(t *testing.T) {
	shell = "sh"
	blockCommands = nil
	getTerminalHeight = func() int {
		return 6
	}

	cases := []struct {
		cmd    string
		stdin  string
		result string
	}{
		{cmd: "echo a", stdin: "", result: "a\n"},
		{cmd: "echo a 1>&2", stdin: "", result: "a\n"},
		{cmd: "grep a", stdin: "a\nb\na\na\n", result: "a\na\na"},
	}
	for _, tc := range cases {
		so := newStdoutViewPane()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		so.execCommand(ctx, tc.cmd, []byte(tc.stdin))
		if !(so.GetText(true) == tc.result) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, so.GetText(true)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestTransform(t *testing.T) {
	cases := []struct {
		line   int
		limit  int
		input  string
		result string
	}{
		{
			line:   0,
			limit:  3,
			input:  "foo\n",
			result: "foo\n",
		},
		{
			line:   0,
			limit:  2,
			input:  "foo\nbar\n",
			result: "foo\nbar",
		},
		{
			line:   0,
			limit:  1,
			input:  "foo\nbar\n",
			result: "foo",
		},
		{
			line:   0,
			limit:  3,
			input:  strings.Repeat("a", 4096) + "foo\n",
			result: strings.Repeat("a", 4096) + "foo\n",
		},
		{
			line:   0,
			limit:  2,
			input:  strings.Repeat("a", 4096) + "foo\nbar\n",
			result: strings.Repeat("a", 4096) + "foo\nbar",
		},
		{
			line:   0,
			limit:  1,
			input:  strings.Repeat("a", 4096) + "foo\nbar\n",
			result: strings.Repeat("a", 4096) + "foo",
		},
		{
			line:   5,
			limit:  5,
			input:  "foo\n",
			result: "",
		},
	}
	for _, tc := range cases {
		stdin := bytes.NewBufferString(tc.input)
		stdout := new(bytes.Buffer)

		tt := &textLineTransformer{
			line:  tc.line,
			limit: tc.limit,
			temp:  []byte(""),
		}
		w := transform.NewWriter(stdout, tt)
		io.Copy(w, stdin)

		if stdout.String() != tc.result {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, stdout.String()), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

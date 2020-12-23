package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"golang.org/x/text/transform"
)

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

	c := newCliPane()
	for _, tc := range cases {
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
	c := newCliPane()
	for _, tc := range cases {
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
	v := newViewPane("")
	for _, tc := range cases {
		v.setData([]byte(tc.input))
		if !(bytes.Equal(v.data, []byte(tc.result))) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, string(v.data)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
		if !(v.GetText(true) == tc.result2) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, v.GetText(true)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result2), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
	}
}

func TestExecCommand(t *testing.T) {
	shell = "sh"
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
	v := newViewPane("test")
	for _, tc := range cases {
		v.execCommand(tc.cmd, []byte(tc.stdin))
		if !(bytes.Equal(v.data, []byte(tc.result))) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, string(v.data)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result), "\n", "\\n", -1)
			t.Errorf("\n%s\n%s", r, e)
		}
		if !(v.GetText(true) == tc.result2) {
			r := strings.Replace(fmt.Sprintf(`result:   "%s"`, v.GetText(true)), "\n", "\\n", -1)
			e := strings.Replace(fmt.Sprintf(`expected: "%s"`, tc.result2), "\n", "\\n", -1)
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

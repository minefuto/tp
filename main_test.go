package main

import (
	"bytes"
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
			t.Error("\nprompt:\n", c.prompt, "\nexpected:\n", tc.prompt, "\n")
		}
		if !(c.GetText() == tc.text) {
			t.Error("\ntext:\n", c.GetText(), "\nexpected:\n", tc.text, "\n")
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
			t.Error("\nprompt:\n", c.prompt, "\nexpected:\n", tc.result, "\n")
		}
	}
}

func TestSetData(t *testing.T) {
	cases := []struct {
		input  string
		result string
	}{
		{input: "a", result: "a"},
	}
	v := newViewPane("")
	for _, tc := range cases {
		v.setData([]byte(tc.input))
		if !(bytes.Equal(v.data, []byte(tc.result))) {
			t.Error("")
		}
	}
}

func TestExecCommand(t *testing.T) {
	cases := []struct {
		cmd    string
		input  string
		result string
	}{
		{cmd: "echo a", input: "", result: "a\n"},
		{cmd: "grep a", input: "a\nb", result: "a\n"},
	}
	v := newViewPane("test")
	for _, tc := range cases {
		v.execCommand(tc.cmd, []byte(tc.input))
		if !(bytes.Equal(v.data, []byte(tc.result))) {
			t.Error("")
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
			t.Error("expected: ", tc.result, ", value: ", stdout.String())
		}
	}
}

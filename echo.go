package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

type cmdEcho struct {
	name        string
	description string
	syntax      string
	re          *regexp.Regexp
}

func newCmdEcho() cmdEcho {
	return cmdEcho{
		syntax:      "!echo message",
		description: "Echo message",
		re:          regexp.MustCompile(`^!echo .+`),
	}
}

func (cmd cmdEcho) Syntax() string {
	return cmd.syntax
}

func (cmd cmdEcho) Description() string {
	return cmd.description
}

func (cmd cmdEcho) Match(text string) bool {
	return cmd.re.MatchString(text)
}

func (cmd cmdEcho) Run(w io.Writer, title, from, text string) error {
	echoText := strings.TrimSpace(strings.TrimPrefix(text, "!echo"))
	fmt.Fprintf(w, "msg %s %s said: %s\n", title, from, echoText)
	return nil
}

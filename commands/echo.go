// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

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

func NewCmdEcho() Command {
	return cmdEcho{
		syntax:      "!e message",
		description: "Echo message",
		re:          regexp.MustCompile(`^!e .+`),
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
	echoText := strings.TrimSpace(strings.TrimPrefix(text, "!e"))
	fmt.Fprintf(w, "msg %s Echo: %s said \"%s\"\n", title, from, echoText)
	return nil
}

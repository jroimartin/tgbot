// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"regexp"
)

type cmdQuotes struct {
	name        string
	description string
	syntax      string
	re          *regexp.Regexp

	// quotesrv config
	endpoint string
	user     string
	passwd   string
}

func newCmdQuotes(endpoint, user, passwd string) cmdQuotes {
	return cmdQuotes{
		syntax:      "!q [message]",
		description: "If message, add a quote. Otherwise, return a random one",
		re:          regexp.MustCompile(`^!q .+`),
	}
}

func (cmd cmdQuotes) Syntax() string {
	return cmd.syntax
}

func (cmd cmdQuotes) Description() string {
	return cmd.description
}

func (cmd cmdQuotes) Match(text string) bool {
	return cmd.re.MatchString(text)
}

func (cmd cmdQuotes) Run(w io.Writer, title, from, text string) error {
	//TODO
	//echoQuote := strings.TrimSpace(strings.TrimPrefix(text, "!q "))
	//fmt.Fprintf(w, "msg %s %s said: %s\n", title, from, echoText)
	return nil
}

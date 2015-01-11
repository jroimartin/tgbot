// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

type cmdQuotes struct {
	name        string
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      QuotesConfig
}

type QuotesConfig struct {
	Enabled  bool
	Endpoint string
	User     string
	Password string
}

func NewCmdQuotes(w io.Writer, config QuotesConfig) Command {
	return &cmdQuotes{
		syntax:      "!q [message]",
		description: "If message, add a quote. Otherwise, return a random one",
		re:          regexp.MustCompile(`^!q($| .+$)`),
		w:           w,
		config:      config,
	}
}

func (cmd *cmdQuotes) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdQuotes) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdQuotes) Description() string {
	return cmd.description
}

func (cmd *cmdQuotes) Match(text string) bool {
	return cmd.re.MatchString(text)
}

func (cmd *cmdQuotes) Run(title, from, text string) error {
	var (
		msg string
		err error
	)

	quoteText := strings.TrimSpace(strings.TrimPrefix(text, "!q"))
	if quoteText == "" {
		msg, err = cmd.randomQuote(title)
	} else {
		msg, err = cmd.addQuote(title, quoteText)
	}
	if err != nil {
		fmt.Fprintf(cmd.w, "msg %v error: cannot get or send quote\n", title)
		return err
	}

	fmt.Fprintf(cmd.w, "msg %v %v\n", title, msg)
	return nil
}

func (cmd *cmdQuotes) Shutdown() error {
	return nil
}

func (cmd *cmdQuotes) randomQuote(title string) (msg string, err error) {
	req, err := http.NewRequest("GET", cmd.config.Endpoint, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(cmd.config.User, cmd.config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot get quote (%v)", res.StatusCode)
	}

	quotes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(quotes), "\n")
	if len(lines) <= 1 { // If there aren't quotes, lines == []string{""}
		return "", errors.New("no quotes")
	}

	rndInt := rand.Intn(len(lines) - 1)
	rndQuote := lines[rndInt]

	return fmt.Sprintf("Random quote: %v\n", rndQuote), nil
}

func (cmd *cmdQuotes) addQuote(title string, text string) (msg string, err error) {
	r := strings.NewReader(text)
	req, err := http.NewRequest("POST", cmd.config.Endpoint, r)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(cmd.config.User, cmd.config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cannot add quote (%v - %v: %v)", res.StatusCode, title, text)
	}

	return fmt.Sprintf("New quote added: %v\n", text), nil
}

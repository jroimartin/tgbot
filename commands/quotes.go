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
	"log"
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
	config      QuotesConfig
}

type QuotesConfig struct {
	Enabled  bool
	Endpoint string
	User     string
	Password string
}

func NewCmdQuotes(config QuotesConfig) Command {
	return &cmdQuotes{
		syntax:      "!q [message]",
		description: "If message, add a quote. Otherwise, return a random one",
		re:          regexp.MustCompile(`^!q($| .+$)`),
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

func (cmd *cmdQuotes) Run(w io.Writer, title, from, text string) error {
	var err error

	quoteText := strings.TrimSpace(strings.TrimPrefix(text, "!q"))
	if quoteText == "" {
		err = cmd.randomQuote(w, title)
	} else {
		err = cmd.addQuote(w, title, quoteText)
	}
	return err
}

func (cmd *cmdQuotes) Shutdown() error {
	return nil
}

func (cmd *cmdQuotes) randomQuote(w io.Writer, title string) error {
	req, err := http.NewRequest("GET", cmd.config.Endpoint, nil)
	if err != nil {
		fmt.Fprintf(w, "msg %v error: Cannot get quote\n", title)
		return err
	}
	req.SetBasicAuth(cmd.config.User, cmd.config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "msg %v error: Cannot get quote\n", title)
		return err
	}

	quotes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Fprintf(w, "msg %v error: Cannot get quote\n", title)
		return err
	}
	lines := strings.Split(string(quotes), "\n")
	if len(lines) <= 1 { // If there aren't quotes, lines == []string{""}
		fmt.Fprintf(w, "msg %v error: no quotes\n", title)
		return errors.New("no quotes")
	}

	rndInt := rand.Intn(len(lines) - 1)
	rndQuote := lines[rndInt]
	log.Println(rndQuote)

	fmt.Fprintf(w, "msg %v Random quote: %v\n", title, rndQuote)
	return nil
}

func (cmd *cmdQuotes) addQuote(w io.Writer, title string, text string) error {
	r := strings.NewReader(text)
	req, err := http.NewRequest("POST", cmd.config.Endpoint, r)
	if err != nil {
		fmt.Fprintf(w, "msg %v error: Cannot add quote\n", title)
		return err
	}
	req.SetBasicAuth(cmd.config.User, cmd.config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "msg %v error: Cannot add quote\n", title)
		return err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "msg %v error (%v): Cannot add quote\n", title, res.StatusCode)
		return fmt.Errorf("cannot add quote (%v - %v: %v)", res.StatusCode, title, text)
	}

	fmt.Fprintf(w, "msg %v New quote added: %v\n", title, text)
	return nil
}

// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
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
	config      quotesConfig
}

type quotesConfig struct {
	Endpoint string
	User     string
	Password string
}

func newCmdQuotes(config quotesConfig) cmdQuotes {
	return cmdQuotes{
		syntax:      "!q [message]",
		description: "If message, add a quote. Otherwise, return a random one",
		re:          regexp.MustCompile(`^!q($| .+$)`),
		config:      config,
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
	var err error

	quoteText := strings.TrimSpace(strings.TrimPrefix(text, "!q"))
	if quoteText == "" {
		err = cmd.randomQuote(w, title)
	} else {
		err = cmd.addQuote(w, title, quoteText)
	}
	return err
}

func (cmd cmdQuotes) randomQuote(w io.Writer, title string) error {
	req, err := http.NewRequest("GET", cmd.config.Endpoint, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(cmd.config.User, cmd.config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	quotes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	lines := strings.Split(string(quotes), "\n")
	rndInt := rand.Intn(len(lines) - 1)
	rndQuote := lines[rndInt]
	log.Println(rndQuote)

	fmt.Fprintf(w, "msg %s %s\n", title, rndQuote)
	return nil
}

func (cmd cmdQuotes) addQuote(w io.Writer, title string, text string) error {
	r := strings.NewReader(text)
	req, err := http.NewRequest("POST", cmd.config.Endpoint, r)
	if err != nil {
		return err
	}
	req.SetBasicAuth(cmd.config.User, cmd.config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "msg %s error (%d): Cannot add quote\n", title, res.StatusCode)
		return fmt.Errorf("Cannot add quote (%d - %s: %s)", res.StatusCode, title, text)
	}

	fmt.Fprintf(w, "msg %s New quote added: %s\n", title, text)
	return nil
}

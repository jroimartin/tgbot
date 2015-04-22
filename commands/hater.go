// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
)

type cmdHater struct {
	description string
	syntax      string
	w           io.Writer
	config      HaterConfig
}

type HaterConfig struct {
	Enabled bool
	Topic   []haterTopic
}

type haterTopic struct {
	Regexp string
	DB     string
}

func NewCmdHater(w io.Writer, config HaterConfig) Command {
	return &cmdHater{
		syntax:      "",
		description: "Topic hater",
		w:           w,
		config:      config,
	}
}

func (cmd *cmdHater) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdHater) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdHater) Description() string {
	return cmd.description
}

func (cmd *cmdHater) Match(text string) bool {
	for _, t := range cmd.config.Topic {
		match, err := regexp.MatchString(t.Regexp, text)
		if err != nil {
			return false
		}
		if match {
			return true
		}
	}
	return false
}

func (cmd *cmdHater) Run(title, from, text string) error {
	var topic haterTopic
	for _, t := range cmd.config.Topic {
		match, err := regexp.MatchString(t.Regexp, text)
		if err != nil {
			return err
		}
		if match {
			topic = t
			break
		}
	}
	if topic.DB == "" {
		return errors.New("text does not match")
	}
	file, err := ioutil.ReadFile(topic.DB)
	if err != nil {
		return err
	}
	lines := strings.Split(string(file), "\n")
	if len(lines) <= 1 {
		return errors.New("empty file")
	}
	rndInt := rand.Intn(len(lines) - 1)
	rndLine := lines[rndInt]
	fmt.Fprintf(cmd.w, "msg %v %v\n", title, rndLine)
	return nil
}

func (cmd *cmdHater) Shutdown() error {
	return nil
}

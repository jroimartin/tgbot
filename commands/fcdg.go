// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/jroimartin/tgbot/utils"
)

const fcdgUrl = "http://lolcathost.org/4cdg/"

var imgRe = regexp.MustCompile(`<img src=(.*?)>`)

type cmdFcdg struct {
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      FcdgConfig

	tempDir string
}

type FcdgConfig struct {
	Enabled bool
}

func NewCmdFcdg(w io.Writer, config FcdgConfig) Command {
	return &cmdFcdg{
		syntax:      "!4",
		description: "return a random card from the 4cdg",
		re:          regexp.MustCompile(`^!4$`),
		w:           w,
		config:      config,
	}
}

func (cmd *cmdFcdg) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdFcdg) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdFcdg) Description() string {
	return cmd.description
}

func (cmd *cmdFcdg) Match(text string) bool {
	return cmd.re.MatchString(text)
}

// Shutdown should remove the temp dir on exit.
func (cmd *cmdFcdg) Shutdown() error {
	if cmd.tempDir == "" {
		return nil
	}
	log.Println("Removing 4cdg pics dir:", cmd.tempDir)
	if err := os.RemoveAll(cmd.tempDir); err != nil {
		return err
	}
	return nil
}

func (cmd *cmdFcdg) Run(title, from, text string) error {
	if cmd.tempDir == "" {
		var err error
		cmd.tempDir, err = ioutil.TempDir("", "tgbot-4cdg-")
		if err != nil {
			fmt.Fprintf(cmd.w, "msg %v error: internal command error\n", title)
			return err
		}
		log.Println("Created 4cdg pics dir:", cmd.tempDir)
	}

	path, err := cmd.randomCard()
	if err != nil {
		fmt.Fprintf(cmd.w, "msg %v error: cannot get pic\n", title)
		return err
	}

	fmt.Fprintf(cmd.w, "send_photo %v %v\n", title, path)
	return nil
}

// getCard returns a random card from the 4cdg
func (cmd *cmdFcdg) randomCard() (filePath string, err error) {
	// Get random pic ID
	resp, err := http.Get(fcdgUrl + "?card")
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %v (%v)", resp.Status, resp.StatusCode)
	}

	matches := imgRe.FindStringSubmatch(string(body))
	if len(matches) != 2 {
		return "", errors.New("regexp error")
	}

	// Download pic
	filePath, err = utils.Download(cmd.tempDir, "", fcdgUrl+matches[1])
	if err != nil {
		return "", err
	}
	return filePath, nil
}

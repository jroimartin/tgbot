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
	"math/rand"
	"os"
	"regexp"
	"strings"

	"github.com/jroimartin/tgbot/utils"
)

type cmdBing struct {
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      BingConfig

	tempDir string
}

type BingConfig struct {
	Enabled bool
	Key     string
	Limit   int
}

func NewCmdBing(w io.Writer, config BingConfig) Command {
	return &cmdBing{
		syntax:      "!sb query",
		description: "Search Bing images by query",
		re:          regexp.MustCompile(`^!sb ([\w ]+)$`),
		w:           w,
		config:      config,
	}
}

func (cmd *cmdBing) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdBing) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdBing) Description() string {
	return cmd.description
}

func (cmd *cmdBing) Match(text string) bool {
	return cmd.re.MatchString(text)
}

// Shutdown should remove the temp dir on exit.
func (cmd *cmdBing) Shutdown() error {
	if cmd.tempDir == "" {
		return nil
	}
	log.Println("Removing Bing pics dir:", cmd.tempDir)
	if err := os.RemoveAll(cmd.tempDir); err != nil {
		return err
	}
	return nil
}

func (cmd *cmdBing) Run(title, from, text string) error {
	var err error

	if cmd.tempDir == "" {
		cmd.tempDir, err = ioutil.TempDir("", "tgbot-bing-")
		if err != nil {
			fmt.Fprintf(cmd.w, "msg %v error: internal command error\n", title)
			return err
		}
		log.Println("Created Bing pics dir:", cmd.tempDir)
	}

	query := strings.TrimSpace(strings.TrimPrefix(text, "!sb"))
	query = strings.Replace(query, " ", "+", -1)
	path, err := cmd.search(query)
	if err != nil {
		fmt.Fprintf(cmd.w, "msg %v error: cannot get pic\n", title)
		return err
	}

	// Send to tg as photo
	fmt.Fprintf(cmd.w, "send_photo %v %v\n", title, path)
	return nil
}

// search returns a pic from Bing after a search using the given query.
func (cmd *cmdBing) search(query string) (filePath string, err error) {
	bs := utils.NewBingSearch(cmd.config.Key)
	if cmd.config.Limit > 0 {
		bs.Limit = cmd.config.Limit
	}

	results, err := bs.Query(utils.Image, query)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", errors.New("no pics")
	}
	rndInt := rand.Intn(len(results))

	// Download pic
	filePath, err = utils.Download(cmd.tempDir, "", results[rndInt].MediaUrl)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

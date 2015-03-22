// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/jroimartin/tgbot/utils"
)

const picsURL = "http://ano.lolcathost.org/pics/"

type cmdAno struct {
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      AnoConfig

	// Regexp used to get the pic URL
	tempDir string
}

type AnoConfig struct {
	Enabled bool
}

func NewCmdAno(w io.Writer, config AnoConfig) Command {
	return &cmdAno{
		syntax:      "!a [tags]",
		description: "if tags, search ANO by tags (comma-separated). Otherwise return a random pic",
		re:          regexp.MustCompile(`^!a($| [\w ,]+$)`),
		w:           w,
		config:      config,
	}
}

func (cmd *cmdAno) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdAno) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdAno) Description() string {
	return cmd.description
}

func (cmd *cmdAno) Match(text string) bool {
	return cmd.re.MatchString(text)
}

// Shutdown should remove the temp dir on exit.
func (cmd *cmdAno) Shutdown() error {
	if cmd.tempDir == "" {
		return nil
	}
	log.Println("Removing ANO pics dir:", cmd.tempDir)
	if err := os.RemoveAll(cmd.tempDir); err != nil {
		return err
	}
	return nil
}

func (cmd *cmdAno) Run(title, from, text string) error {
	var (
		path string
		err  error
	)

	if cmd.tempDir == "" {
		cmd.tempDir, err = ioutil.TempDir("", "tgbot-ano-")
		if err != nil {
			fmt.Fprintf(cmd.w, "msg %v error: internal command error\n", title)
			return err
		}
		log.Println("Created ANO pics dir:", cmd.tempDir)
	}

	tags := strings.TrimSpace(strings.TrimPrefix(text, "!a"))
	if tags == "" {
		path, err = cmd.randomPic(title)
	} else {
		path, err = cmd.searchTag(title, strings.Split(tags, ","))
	}
	if err != nil {
		fmt.Fprintf(cmd.w, "msg %v error: cannot get pic\n", title)
		return err
	}

	// Send to tg as photo
	fmt.Fprintf(cmd.w, "msg %v What has been seen cannot be unseen...\n", title)
	fmt.Fprintf(cmd.w, "send_photo %v %v\n", title, path)
	return nil
}

// randomPic returns a random pic from ANO
func (cmd *cmdAno) randomPic(title string) (filePath string, err error) {
	var data struct {
		Pic struct {
			ID string
		}
	}
	client := &http.Client{}

	// Get random pic ID
	methodRandom := strings.NewReader(`{ "method" : "random" }`)
	req, err := http.NewRequest("POST", "http://ano.lolcathost.org/json/pic.json", methodRandom)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %v (%v)", res.Status, res.StatusCode)
	}

	// Receive data
	bdata, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bdata, &data)
	if err != nil {
		return "", err
	}

	// Download pic
	filePath, err = utils.Download(cmd.tempDir, "", picsURL+data.Pic.ID)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// searchTag returns a pic from ANO with a given tag.
func (cmd *cmdAno) searchTag(title string, tags []string) (filePath string, err error) {
	var data struct {
		Pics []struct {
			ID string
		}
	}
	client := &http.Client{}

	// Get random pic ID
	searchStr := fmt.Sprintf("{ \"method\" : \"searchRelated\", \"tags\" : [%v], \"limit\" : 10 }",
		tagsString(tags))

	methodSearch := strings.NewReader(searchStr)
	req, err := http.NewRequest("POST", "http://ano.lolcathost.org/json/tag.json", methodSearch)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %v (%v)", res.Status, res.StatusCode)
	}

	// Receive data
	bdata, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bdata, &data)
	if err != nil {
		return "", err
	}
	if len(data.Pics) <= 1 {
		return "", errors.New("no pics")
	}

	rndInt := rand.Intn(len(data.Pics) - 1)
	rndData := data.Pics[rndInt]

	// Download pic
	filePath, err = utils.Download(cmd.tempDir, "", picsURL+rndData.ID)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// tagsString construct a valid string with the tags to be
// send to ANO's searchRelated method.
func tagsString(tags []string) string {
	for i := range tags {
		tags[i] = fmt.Sprintf("\"%v\"", strings.TrimSpace(tags[i]))
	}
	return strings.Join(tags, ",")
}

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
	"path"
	"regexp"
	"strings"
)

var errorNew = errors.New("new file")

type cmdAno struct {
	name        string
	description string
	syntax      string
	re          *regexp.Regexp
	config      AnoConfig

	// Regexp used to get the pic URL
	tempDir string
}

type AnoConfig struct {
	Enabled bool
}

func NewCmdAno(config AnoConfig) Command {
	return &cmdAno{
		syntax:      "!a [tags]",
		description: "if tags, search ANO by tags (comma-separated). Otherwise return a random pic",
		re:          regexp.MustCompile(`^!a($| .+$)`),
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

func (cmd *cmdAno) Run(w io.Writer, title, from, text string) error {
	var (
		path string
		err  error
	)

	tags := strings.TrimSpace(strings.TrimPrefix(text, "!a"))
	if tags == "" {
		path, err = cmd.randomPic(title)
	} else {
		path, err = cmd.searchTag(title, strings.Split(tags, ","))
	}
	if err != nil {
		fmt.Fprintf(w, "msg %v error: Cannot get pic\n", title)
		return err
	}

	// Send to tg as document
	fmt.Fprintf(w, "msg %v What has been seen cannot be unseen...\n", title)
	fmt.Fprintf(w, "send_document %v %v\n", title, path)
	return nil
}

// randomPic returns a random pic from ANO
func (cmd *cmdAno) randomPic(title string) (path string, err error) {
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

	// Receive data
	bdata, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bdata, &data)
	if err != nil {
		return "", err
	}

	// Download pic
	path, err = cmd.download(data.Pic.ID)
	if err != nil {
		return "", fmt.Errorf("cannot download pic")
	}
	return path, nil
}

// searchTag returns a pic from ANO with a given tag.
func (cmd *cmdAno) searchTag(title string, tags []string) (path string, err error) {
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

	// Receive data
	bdata, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bdata, &data)
	if err != nil {
		return "", err
	}

	rndInt := rand.Intn(len(data.Pics) - 1)
	rndData := data.Pics[rndInt]

	// Download pic
	path, err = cmd.download(rndData.ID)
	if err != nil {
		return "", fmt.Errorf("cannot download pic")
	}
	return path, nil
}

// tagsString construct a valid string with the tags to be
// send to ANO's searchRelated method.
func tagsString(tags []string) string {
	for i := range tags {
		tags[i] = fmt.Sprintf("\"%v\"", strings.TrimSpace(tags[i]))
	}
	return strings.Join(tags, ",")
}

// download downloads the pic on the given URL and return
// the file.
func (cmd *cmdAno) download(picID string) (path string, err error) {
	res, err := http.Get("http://ano.lolcathost.org/pics/" + picID)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	f, err := cmd.createTempFile(picID)
	if err != nil && err != errorNew {
		return "", err
	}
	defer f.Close()

	if err == errorNew {
		_, err = io.Copy(f, res.Body)
		if err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}

// createTempFile tries to create a new temporary file with
// the given name. It is important to note that the error
// will be errorNew if the file did not exist. It also
// creates a temporary directory in case it has not been
// created yet.
func (cmd *cmdAno) createTempFile(filename string) (*os.File, error) {
	var err, ferr error

	if cmd.tempDir == "" {
		cmd.tempDir, err = ioutil.TempDir("", "tgbot")
		if err != nil {
			return nil, err
		}
	}

	path := path.Join(cmd.tempDir, filename)

	var f *os.File
	f, err = os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) {
		f, err = os.Create(path)
		if err != nil {
			return nil, err
		}
		ferr = errorNew
	}

	return f, ferr
}

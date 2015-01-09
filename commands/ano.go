// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
)

type cmdAno struct {
	name        string
	description string
	syntax      string
	re          *regexp.Regexp
	config      AnoConfig

	// Regexp used to get the pic URL
	picRegexp *regexp.Regexp
	picsDir   string
}

type AnoConfig struct {
	Enabled bool
}

func NewCmdAno(config AnoConfig) Command {
	return &cmdAno{
		syntax:      "!a",
		description: "Return a random ANO pic",
		re:          regexp.MustCompile(`^!a$`),
		config:      config,
		picRegexp:   regexp.MustCompile(`/pics/[^"]+`),
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

func (cmd *cmdAno) Shutdown() error {
	if cmd.picsDir == "" {
		return nil
	}
	log.Println("Removing ANO pics dir:", cmd.picsDir)
	if err := os.RemoveAll(cmd.picsDir); err != nil {
		return err
	}
	return nil
}

func (cmd *cmdAno) Run(w io.Writer, title, from, text string) error {
	// Get pic URL
	res, err := getResponse("http://ano.lolcathost.org/random.mhtml")
	if err != nil {
		fmt.Fprintf(w, "msg %s error: Cannot get pic\n", title)
		return fmt.Errorf("cannot get pic. failed request to random page")
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	picPath := cmd.picRegexp.FindString(string(body))
	if picPath == "" {
		fmt.Fprintf(w, "msg %s error: Cannot get pic\n", title)
		return fmt.Errorf("cannot get pic. regexp didn't match")
	}
	picURL := "http://ano.lolcathost.org" + picPath

	// Download pic
	res, err = getResponse(picURL)
	if err != nil {
		fmt.Fprintf(w, "msg %s error: Cannot get pic\n", title)
		return fmt.Errorf("cannot get pic. failed request to pic URL")
	}

	// Create tmp file
	log.Println(cmd.picsDir)
	if cmd.picsDir == "" {
		cmd.picsDir, err = ioutil.TempDir("", "tgbot")
		if err != nil {
			fmt.Fprintf(w, "msg %s error: Cannot get pic\n", title)
			return err
		}
	}

	var picFile *os.File
	picTmpPath := path.Join(cmd.picsDir, path.Base(picPath))
	picFile, err = os.Open(picTmpPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		picFile, err = os.Create(picTmpPath)
		if err != nil {
			fmt.Fprintf(w, "msg %s error: Cannot get pic\n", title)
			return err
		}
	}
	defer picFile.Close()

	_, err = io.Copy(picFile, res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Fprintf(w, "msg %s error: Cannot get pic\n", title)
		return err
	}

	// Send to tg as document
	fmt.Fprintf(w, "msg %s What has been seen cannot be unseen... %s\n", title, picURL)
	fmt.Fprintf(w, "send_document %s %s\n", title, picFile.Name())
	return nil
}

func getResponse(URL string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{
		Name:  "ANO_PREF",
		Value: "ip%3D20%26lt%3D0%26ic%3D2%26ml%3D0%26sn%3D0%26it%3D0%26ns%3D0%26md%3D0",
	})
	return client.Do(req)
}

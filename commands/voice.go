// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"regexp"
)

type cmdVoice struct {
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      VoiceConfig

	tempDir string
}

type VoiceConfig struct {
	Enabled bool
}

func NewCmdVoice(w io.Writer, config VoiceConfig) Command {
	return &cmdVoice{
		syntax:      "!v[en|es|fr] message",
		description: "text to speech generator courtesy of google translate",
		re:          regexp.MustCompile(`^!v(es|en|fr)? (.+$)`),
		w:           w,
		config:      config,
	}
}

func (cmd *cmdVoice) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdVoice) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdVoice) Description() string {
	return cmd.description
}

func (cmd *cmdVoice) Match(text string) bool {
	return cmd.re.MatchString(text)
}

// Shutdown should remove the temp dir on exit.
func (cmd *cmdVoice) Shutdown() error {
	if cmd.tempDir == "" {
		return nil
	}
	log.Println("Removing VOICE sounds dir:", cmd.tempDir)
	if err := os.RemoveAll(cmd.tempDir); err != nil {
		return err
	}
	return nil
}

func (cmd *cmdVoice) Run(title, from, text string) error {
	var (
		path string
		err  error
	)

	if cmd.tempDir == "" {
		cmd.tempDir, err = ioutil.TempDir("", "tgbot-voice-")
		if err != nil {
			fmt.Fprintf(cmd.w, "msg %v error: internal command error\n", title)
			return err
		}
		log.Println("Created VOICE sounds dir:", cmd.tempDir)
	}

	// Get language and text
	matches := cmd.re.FindStringSubmatch(text)
	lang := matches[1]
	msg := matches[2]

	// Download sound 
	path, err = download(cmd.tempDir, ".mp3", setResourceUrl(lang, msg))

	if err != nil {
		fmt.Fprintf(cmd.w, "msg %v error: cannot get sound\n", title)
		return err
	}

	// Send to tg as document
	fmt.Fprintf(cmd.w, "send_document %v %v\n", title, path)
	return nil
}

func setResourceUrl(lang, text string) string {
	const gooTrans = "http://translate.google.com/translate_tts"
	if lang == "" {
		lang = "es"
	}
	return gooTrans + "?tl=" + url.QueryEscape(lang) + "&q=" + url.QueryEscape(text)
}

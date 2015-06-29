// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ChimeraCoder/anaconda"
)

type cmdTweet struct {
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      TweetConfig
}

type TweetConfig struct {
	Enabled           bool
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func NewCmdTweet(w io.Writer, config TweetConfig) Command {
	return &cmdTweet{
		syntax:      "!tw tweet",
		description: "Tweet a message",
		re:          regexp.MustCompile(`^!tw .+`),
		w:           w,
		config:      config,
	}
}

func (cmd *cmdTweet) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdTweet) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdTweet) Description() string {
	return cmd.description
}

func (cmd *cmdTweet) Match(text string) bool {
	return cmd.re.MatchString(text)
}

func (cmd *cmdTweet) Run(title, from, text string) error {
	tweetText := strings.TrimSpace(strings.TrimPrefix(text, "!tw"))

	anaconda.SetConsumerKey(cmd.config.ConsumerKey)
	anaconda.SetConsumerSecret(cmd.config.ConsumerSecret)
	api := anaconda.NewTwitterApi(cmd.config.AccessToken, cmd.config.AccessTokenSecret)

	if tweetLen := len(tweetText); tweetLen > 140 {
		fmt.Fprintf(cmd.w, "msg %v %v chars? Mmm to much for me, size actually matters\n", title, tweetLen)
		return errors.New("invalid message length")
	} else {
		if _, err := api.PostTweet(tweetText, nil); err != nil {
			fmt.Fprintf(cmd.w, "msg %v Useless humans...something went wrong\n", title)
			return err
		}
		fmt.Fprintf(cmd.w, "msg %v Congrats you did it, new boring tweet posted\n", title)
	}
	return nil
}

func (cmd *cmdTweet) Shutdown() error {
	return nil
}

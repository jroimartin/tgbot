// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type cmdBreakfast struct {
	description string
	syntax      string
	re          *regexp.Regexp
	w           io.Writer
	config      BreakfastConfig

	// Stored items
	items map[string][]string
}

type BreakfastConfig struct {
	Enabled bool
}

func NewCmdBreakfast(w io.Writer, config BreakfastConfig) Command {
	return &cmdBreakfast{
		syntax: "!b[-] [item]",
		description: "If item, add a item to the list. Otherwise, return the list. " +
			"!b- [n]: If n, remove item n. Otherwise, reset list.",
		re:     regexp.MustCompile(`^!b(($| [^\r\n]+$)|(-$|- \d+$))`),
		w:      w,
		config: config,
		items:  make(map[string][]string),
	}
}

func (cmd *cmdBreakfast) Enabled() bool {
	return cmd.config.Enabled
}

func (cmd *cmdBreakfast) Syntax() string {
	return cmd.syntax
}

func (cmd *cmdBreakfast) Description() string {
	return cmd.description
}

func (cmd *cmdBreakfast) Match(text string) bool {
	return cmd.re.MatchString(text)
}

func (cmd *cmdBreakfast) Shutdown() error {
	return nil
}

func (cmd *cmdBreakfast) Run(title, from, text string) error {
	var err error

	if strings.HasPrefix(text, "!b-") {
		bfText := strings.TrimSpace(strings.TrimPrefix(text, "!b-"))
		if bfText == "" {
			// !b-: Reset list
			err = cmd.listReset(title)
		} else {
			// !b- n: Remove item n
			err = cmd.removeItem(title, bfText)
		}
	} else {
		bfText := strings.TrimSpace(strings.TrimPrefix(text, "!b"))
		if bfText == "" {
			// !b: List
			err = cmd.listItems(title)
		} else {
			// !b item: Add item to the list
			err = cmd.addItem(title, from, bfText)
		}
	}
	if err != nil {
		fmt.Fprintf(cmd.w, "msg %v error: cannot get or add items\n", title)
		return err
	}
	return nil
}

func (cmd *cmdBreakfast) addItem(title, from, text string) error {
	item := fmt.Sprintf("%v: %v", from, text)
	cmd.items[title] = append(cmd.items[title], item)
	fmt.Fprintf(cmd.w, "msg %v New item added: \"%v\"\n", title, item)
	return nil
}

func (cmd *cmdBreakfast) listItems(title string) error {
	items, ok := cmd.items[title]
	if !ok || len(items) < 1 {
		return errors.New("no items")
	}

	for i, item := range items {
		fmt.Fprintf(cmd.w, "msg %v [%v] %v\n", title, i, item)
	}

	return nil
}

func (cmd *cmdBreakfast) listReset(title string) error {
	delete(cmd.items, title)
	fmt.Fprintf(cmd.w, "msg %v The list has been reset\n", title)
	return nil
}

func (cmd *cmdBreakfast) removeItem(title, text string) error {
	n, err := strconv.Atoi(text)
	if err != nil {
		return err
	}

	items, ok := cmd.items[title]
	if !ok {
		return errors.New("list not found")
	}
	if n < 0 || n > len(items)-1 {
		return errors.New("n is out of bounds")
	}

	cmd.items[title] = append(items[:n], items[n+1:]...)
	fmt.Fprintf(cmd.w, "msg %v The item %v has been removed\n", title, n)

	return nil
}

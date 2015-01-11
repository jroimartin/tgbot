// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/jroimartin/tgbot/commands"
)

var (
	// Message format: "[MSG] title from msg".
	msgRegexp = regexp.MustCompile(`^\[MSG\] ([^ ]+) ([^ ]+) (.*)$`)

	// Global configuration.
	globalConfig config

	// Enabled commands.
	enabledCommands = []commands.Command{}

	// Channel used to receive OS signals.
	sig = make(chan os.Signal, 1)

	// Communication pipes with the tg client
	stdoutTg io.ReadCloser
	stdinTg  io.WriteCloser
)

// Configuration used for bot and commands.
type config struct {
	TgBin     string
	TgPubKey  string
	MinOutput string
	Chat      string
	Echo      commands.EchoConfig
	Quotes    commands.QuotesConfig
	Ano       commands.AnoConfig
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tgbot config")
		os.Exit(2)
	}
	configFile := os.Args[1]
	if _, err := toml.DecodeFile(configFile, &globalConfig); err != nil {
		log.Fatalln(err)
	}
	globalConfig.Chat = strings.Replace(globalConfig.Chat, " ", "_", -1)

	// Clean shutdown with Ctrl-C
	signal.Notify(sig, os.Interrupt, os.Kill)

	if err := listenAndServe(); err != nil {
		log.Fatalln(err)
	}

	log.Println("Bye!")
}

func listenAndServe() error {
	// -R: disable readline, -C: disable color, -D: disable output, -s: lua script
	cmd := exec.Command(globalConfig.TgBin, "-R", "-C", "-D",
		"-s", globalConfig.MinOutput,
		"-k", globalConfig.TgPubKey)

	stdoutTg, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdinTg, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// initCommads must be caled after stdinTg and stdoutTg has bee initialized
	initCommads()
	defer shutdownCommands()

	log.Println("Monitoring...")
	s := bufio.NewScanner(stdoutTg)
readLoop:
	for {
		select {
		case <-sig: // Ctrl-C
			break readLoop
		default:
			if !s.Scan() {
				break readLoop
			}
			handleMsg(s.Text())
		}
	}
	if err := s.Err(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// initCommads enables plugins.
func initCommads() {
	enabledCommands = append(enabledCommands, commands.NewCmdEcho(stdinTg, globalConfig.Echo))
	enabledCommands = append(enabledCommands, commands.NewCmdQuotes(stdinTg, globalConfig.Quotes))
	enabledCommands = append(enabledCommands, commands.NewCmdAno(stdinTg, globalConfig.Ano))
}

// shutdownCommands gracefully shuts down all commands.
func shutdownCommands() {
	for _, cmd := range enabledCommands {
		if !cmd.Enabled() {
			continue
		}
		if err := cmd.Shutdown(); err != nil {
			log.Println(err)
		}
	}
}

// handleMsg parses the message and calls handleCommand
// with the title, from and text of the message.
func handleMsg(msg string) {
	sm := msgRegexp.FindStringSubmatch(msg)
	if len(sm) != 4 {
		return
	}
	title := sm[1]
	from := sm[2]
	text := sm[3]
	log.Printf("DEBUG: title=%v, from=%v, text=%v\n", title, from, text)

	if !isMonitored(title) {
		return
	}

	handleCommand(title, from, text)
}

// isMonitored returns true if "title" is monitored.
func isMonitored(title string) bool {
	if globalConfig.Chat == "" || globalConfig.Chat == title {
		return true
	}
	return false
}

// handleCommand selects the command and executes it.
func handleCommand(title, from, text string) {
	if strings.HasPrefix(text, "!?") {
		for _, cmd := range enabledCommands {
			if cmd.Enabled() {
				fmt.Fprintf(stdinTg, "msg %v - %v: %v\n", title, cmd.Syntax(), cmd.Description())
			}
		}
		return
	}

	for _, cmd := range enabledCommands {
		if cmd.Enabled() && cmd.Match(text) {
			if err := cmd.Run(title, from, text); err != nil {
				log.Println(err)
			}
			return
		}
	}
}

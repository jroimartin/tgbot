package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
)

var (
	tgbin  = flag.String("tgbin", "", "telegram-cli executable")
	pubkey = flag.String("pubkey", "", "telegram server public key")
	chat   = flag.String("chat", "", "monitored chat (all if not defined")
)

func main() {
	flag.Parse()

	if *tgbin == "" || *pubkey == "" {
		fmt.Fprintln(os.Stderr, "error: -tgbin and -pubkey are mandotory")
		os.Exit(1)
	}

	// Clean shutdown with Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// -R: disable readline, -C: disable color
	cmd := exec.Command(*tgbin, "-R", "-C", "-k", *pubkey)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalln(err)
	}

	// TODO (jrm): multiple commands. !echo is just a POC
	// Format: "[00:00]  chat nick >>> msg"
	re := regexp.MustCompile(`^\[\d{2}:\d{2}\]  (.*?) (.*?) >>> !echo (.*)$`)

	log.Println("Monitoring...")
	s := bufio.NewScanner(stdoutPipe)
readLoop:
	for {
		select {
		case <-c: // Ctrl-C
			break readLoop
		default:
			if !s.Scan() {
				break readLoop
			}
			line := s.Text()
			log.Println("DEBUG:", line)
			sm := re.FindStringSubmatch(line)
			if len(sm) == 4 && (*chat == "" || *chat == sm[1]) {
				fmt.Fprintf(stdinPipe, "msg %s Auto: %s\n", sm[1], sm[3])
			}
		}
	}
	if err := s.Err(); err != nil {
		log.Fatalln(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalln(err)
	}

	log.Println("Bye!")
}

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

// TODO (jrm): Support multiple channels
var chat = flag.String("chat", "", "monitored chat (all if not defined)")

func usage() {
	fmt.Fprintln(os.Stderr, "usage: tgbot [flag] tgbin pubkey minoutput")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 3 {
		usage()
		os.Exit(2)
	}

	tgbin := flag.Arg(0)
	pubkey := flag.Arg(1)
	minoutput := flag.Arg(2)

	// Clean shutdown with Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// -R: disable readline, -C: disable color, -D: disable output, -s: lua script
	cmd := exec.Command(tgbin, "-R", "-C", "-D", "-s", minoutput, "-k", pubkey)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	/*
		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalln(err)
		}
	*/

	if err := cmd.Start(); err != nil {
		log.Fatalln(err)
	}

	// TODO (jrm): multiple commands. !echo is just a POC
	// Format: "title from msg"
	re := regexp.MustCompile(`^\[MSG\] ([^ ]+) ([^ ]+) (.*)$`)

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
			//log.Println("DEBUG:", line)
			sm := re.FindStringSubmatch(line)
			if len(sm) != 4 {
				continue
			}
			title := sm[1]
			from := sm[2]
			msg := sm[3]
			log.Printf("DEBUG: title=%s, from=%s, msg=%s\n", title, from, msg)
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

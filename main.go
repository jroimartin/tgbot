package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
)

var (
	chat = flag.String("chat", "", "monitored chat (all if not defined)")

	// Message format: "[MSG] title from msg"
	msgRegexp = regexp.MustCompile(`^\[MSG\] ([^ ]+) ([^ ]+) (.*)$`)
)

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
	*chat = strings.Replace(*chat, " ", "_", -1)

	// Clean shutdown with Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// -R: disable readline, -C: disable color, -D: disable output, -s: lua script
	cmd := exec.Command(tgbin, "-R", "-C", "-D", "-s", minoutput, "-k", pubkey)

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
			handleMsg(stdinPipe, s.Text())
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

func handleMsg(w io.Writer, msg string) {
	sm := msgRegexp.FindStringSubmatch(msg)
	if len(sm) != 4 {
		return
	}
	title := sm[1]
	from := sm[2]
	text := sm[3]
	log.Printf("DEBUG: title=%s, from=%s, text=%s\n", title, from, text)

	if !isMonitored(title) {
		return
	}

	handleCommand(w, title, from, text)
}

func isMonitored(title string) bool {
	if *chat == "" || *chat == title {
		return true
	}
	return false
}

func handleCommand(w io.Writer, title, from, text string) {
	if strings.HasPrefix(text, "!echo") {
		fmt.Fprintf(w, "msg %s %s\n", title, strings.TrimSpace(strings.TrimPrefix(text, "!echo")))
	}
}

package main

import "io"

type Command interface {
	Syntax() string
	Description() string
	Match(text string) bool
	Run(w io.Writer, title, from, text string) error
}

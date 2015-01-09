// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import "io"

type Command interface {
	Enabled() bool
	Syntax() string
	Description() string
	Match(text string) bool
	Run(w io.Writer, title, from, text string) error
	Shutdown() error
}

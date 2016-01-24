// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package writ_test

import (
	"errors"
	"github.com/ziuchkovski/writ"
	"os"
)

type GoBox struct {
	Link Link `command:"ln" alias:"link" description:"Create a soft or hard link"`
	List List `command:"ls" alias:"list" description:"List directory contents"`
}

type Link struct {
	HelpFlag bool `flag:"h, help" description:"Display this message and exit"`
	Symlink  bool `flag:"s" description:"Create a symlink instead of a hard link"`
}

type List struct {
	HelpFlag   bool `flag:"h, help" description:"Display this message and exit"`
	LongFormat bool `flag:"l" description:"Use long-format output"`
}

func (r *GoBox) Run(p writ.Path, positional []string) {
	// The user didn't specify a subcommand.  Give them help.
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (r *Link) Run(p writ.Path, positional []string) {
	if r.HelpFlag {
		p.Last().ExitHelp(nil)
	}
	if len(positional) != 2 {
		p.Last().ExitHelp(errors.New("ln requires two arguments, OLD and NEW"))
	}
	// Link operation omitted for brevity.  This would be os.Link or os.Symlink
	// based on the r.Symlink value.
}

func (r *List) Run(p writ.Path, positional []string) {
	if r.HelpFlag {
		p.Last().ExitHelp(nil)
	}
	// Listing operation omitted for brevity.  This would be a call to ioutil.ReadDir
	// followed by conditional formatting based on the r.LongFormat value.
}

// This example demonstrates subcommands in a busybox style.  There's no requirement
// that subcommands implement this particular Run() method, or any Run() method at
// at all.
func Example_subcommand() {
	gobox := &GoBox{}
	cmd := writ.New("gobox", gobox)
	cmd.Help.Usage = "Usage: gobox COMMAND [OPTION]... [ARG]..."
	cmd.Subcommand("ln").Help.Usage = "Usage: gobox ln [-s] OLD NEW"
	cmd.Subcommand("ls").Help.Usage = "Usage: gobox ls [-l] [PATH]..."

	path, positional, err := cmd.Decode(os.Args[1:])
	if err != nil {
		cmd.ExitHelp(err)
	}

	// At this point, cmd.Decode() has already decoded option values into the gobox
	// struct, including subcommand values.  We just need to dispatch the command.
	// path.String() is guaranteed to represent the user command selection.
	switch path.String() {
	case "gobox":
		gobox.Run(path, positional)
	case "gobox ln":
		gobox.Link.Run(path, positional)
	case "gobox ls":
		gobox.List.Run(path, positional)
	default:
		panic("BUG: Someone added a new command and forgot to add it's path here")
	}
}

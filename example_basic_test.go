// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package writ_test

import (
	"fmt"
	"github.com/ziuchkovski/writ"
	"strings"
)

type Greeter struct {
	HelpFlag  bool   `flag:"help" description:"Display this help message and exit"`
	Verbosity int    `flag:"v, verbose" description:"Display verbose output"`
	Name      string `option:"n, name" default:"Everyone" description:"The person or people to greet"`
}

// This example uses writ.New() to build a command from the Greeter's
// struct fields.  The resulting *writ.Command decodes and updates the
// Greeter's fields in-place.  The Command.ExitHelp() method is used to
// display help content if --help is specified, or if invalid input
// arguments are received.
func Example_basic() {
	greeter := &Greeter{}
	cmd := writ.New("greeter", greeter)
	cmd.Help.Usage = "Usage: greeter [OPTION]... MESSAGE"
	cmd.Help.Header = "Greet users, displaying MESSAGE"

	// Use cmd.Decode(os.Args[1:]) in a real application
	_, positional, err := cmd.Decode([]string{"-vvv", "--name", "Sam", "How's it going?"})
	if err != nil || greeter.HelpFlag {
		cmd.ExitHelp(err)
	}

	message := strings.Join(positional, " ")
	fmt.Printf("Hi %s! %s\n", greeter.Name, message)
	if greeter.Verbosity > 0 {
		fmt.Printf("I'm feeling re%slly chatty today!\n", strings.Repeat("a", greeter.Verbosity))
	}

	// Output:
	// Hi Sam! How's it going?
	// I'm feeling reaaally chatty today!

	// Help Output:
	// Usage: greeter [OPTION]... MESSAGE
	// Greet users, displaying MESSAGE
	//
	// Available Options:
	//   --help                    Display this help message and exit
	//   -v, --verbose             Display verbose output
	//   -n, --name=ARG            The person or people to greet
}

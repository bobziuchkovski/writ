// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package writ_test

import (
	"fmt"
	"github.com/ziuchkovski/writ"
	"strings"
)

// The Greeter's field tags are parsed into flags and options by writ.New()
type Greeter struct {
	HelpFlag  bool   `flag:"help" description:"display this help message"`
	Verbosity int    `flag:"v, verbose" description:"display verbose output"`
	Name      string `option:"n, name" default:"Everyone" description:"the person to greet"`
}

func Example_basic() {
	// First, the Greeter is parsed into a *Command by writ.New()
	greeter := &Greeter{}
	cmd := writ.New("greeter", greeter)

	// Next, the input arguments are decoded.
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
}

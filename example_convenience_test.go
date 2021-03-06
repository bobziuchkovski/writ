// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package writ_test

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bobziuchkovski/writ"
	"io"
	"os"
	"strings"
)

type ReplacerCmd struct {
	Input        io.Reader         `option:"i" description:"Read input values from FILE (default: stdin)" default:"-" placeholder:"FILE"`
	Output       io.WriteCloser    `option:"o" description:"Write output to FILE (default: stdout)" default:"-" placeholder:"FILE"`
	Replacements map[string]string `option:"r, replace" description:"Replace occurrences of ORIG with NEW" placeholder:"ORIG=NEW"`
	HelpFlag     bool              `flag:"h, help" description:"Display this help text and exit"`
}

// This example demonstrates some of the convenience features offered by writ.
// It uses writ's support for io types and default values to ensure the Input
// and Output fields are initialized.  These default to stdin and stdout due
// to the default:"-" field tags.  The user may specify -i or -o to read from
// or write to a file.  Similarly, the Replacements map is initialized with
// key=value pairs for every -r/--replace option the user specifies.
func Example_convenience() {
	replacer := &ReplacerCmd{}
	cmd := writ.New("replacer", replacer)
	cmd.Help.Usage = "Usage: replacer [OPTION]..."
	cmd.Help.Header = "Perform text replacement according to the -r/--replace option"
	cmd.Help.Footer = "By default, replacer reads from stdin and write to stdout.  Use the -i and -o options to override."

	// Decode input arguments
	_, positional, err := cmd.Decode(os.Args[1:])
	if err != nil || replacer.HelpFlag {
		cmd.ExitHelp(err)
	}
	if len(positional) > 0 {
		cmd.ExitHelp(errors.New("replacer does not accept positional arguments"))
	}

	// At this point, the ReplacerCmd's Input, Output, and Replacements fields are all
	// known-valid and initialized, so we can run the replacement.
	err = replacer.Replace()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// The Replace() method performs the input/output replacements, but is
// not relevant to the example itself.
func (r ReplacerCmd) Replace() error {
	var pairs []string
	for k, v := range r.Replacements {
		pairs = append(pairs, k, v)
	}
	replacer := strings.NewReplacer(pairs...)
	scanner := bufio.NewScanner(r.Input)
	for scanner.Scan() {
		line := scanner.Text()
		_, err := io.WriteString(r.Output, replacer.Replace(line)+"\n")
		if err != nil {
			return err
		}
	}
	err := scanner.Err()
	if err != nil {
		return err
	}
	return r.Output.Close()

	// Help Output:
	// Usage: replacer [OPTION]...
	// Perform text replacement according to the -r/--replace option
	//
	// Available Options:
	//   -i FILE                   Read input values from FILE (default: stdin)
	//   -o FILE                   Write output to FILE (default: stdout)
	//   -r, --replace=ORIG=NEW    Replace occurrences of ORIG with NEW
	//   -h, --help                Display this help text and exit
	//
	// By default, replacer reads from stdin and write to stdout.  Use the -i and -o options to override.
}

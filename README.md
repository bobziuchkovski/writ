[![Build Status](https://travis-ci.org/ziuchkovski/writ.svg?branch=master)](https://travis-ci.org/ziuchkovski/writ)
[![Report Card](http://goreportcard.com/badge/ziuchkovski/writ)](http://goreportcard.com/report/ziuchkovski/writ)
[![Coverage](http://gocover.io/_badge/github.com/ziuchkovski/writ?1)](http://gocover.io/github.com/ziuchkovski/writ)
[![GoDoc](https://godoc.org/github.com/ziuchkovski/writ?status.svg)](https://godoc.org/github.com/ziuchkovski/writ)

# Writ

## Overview

Writ implements command line decoding according to [GNU getopt_long conventions](http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html).  All long and short-form option variations are supported: `--with-x`, `--name Sam`, `--day=Friday`, `-i FILE`, `-vvv`, etc.

Additionally, writ supports subcommands, customizable help output generation, and default values. However, writ is purely a decoder package. Command dispatch and execution are intentionally omitted.

Commands and options may be defined either implicitly via struct tags, or explicitly via direct writ.Command{} and writ.Option{} creation.

## Current version

0.8.4

## Stability and API Promise

Writ is new but stable.  It has thorough test coverage, particularly for command and option parsing.

The API is mostly established, but might change in minor breaking ways prior to the 1.0 release.  Any API changes after 1.0 are guaranteed to remain backwards compatible.  This is similar to the Go language promise.

## Usage

The following examples are copied from writ's package documentation.  Please read the [godocs](https://godoc.org/github.com/ziuchkovski/writ) for additional information.

### Basic Use

```go
package main

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

func main() {
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
```

### Convenience Features

```go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "github.com/ziuchkovski/writ"
    "io"
    "os"
    "strings"
)

// This example demonstrates some of the convenience features offered by writ
// It replaces user-specified words in an input file and writes the results to
// an output file.  By default, the input is read from os.Stdin and written to
// os.Stdout.
type ReplacerCmd struct {
    Input        io.Reader         `option:"i" description:"Read input values from FILE (default: stdin)" default:"-" placeholder:"FILE"`
    Output       io.WriteCloser    `option:"o" description:"Write rendered output to FILE (default: stdout)" default:"-" placeholder:"FILE"`
    Replacements map[string]string `option:"r, replace" description:"Replace occurrences of ORIG with NEW" placeholder:"ORIG=NEW"`
    HelpFlag     bool              `flag:"h, help" description:"Display this help text and exit"`
}

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
}

func main() {
    // Construct the command
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
    // known-valid, so we can run the replacement.
    err = replacer.Replace()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(1)
    }
}
```

### Explicit Commands and Options

```go
// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/ziuchkovski/writ"
	"os"
	"runtime"
)

type Config struct {
	help      bool
	verbosity int

	// A hidden flag
	tolerance float32

	// A dynamically added option for Mac OS
	useQuartz bool
}

// This example demonstrates explicit Command and Option construction
// without the use of writ.New()
func main() {
	config := &Config{}
	cmd := &writ.Command{Name: "explicit"}
	cmd.Options = []*writ.Option{
		{
			Names:       []string{"h", "help"},
			Description: "Display this help text and exit",
			Decoder:     writ.NewFlagDecoder(&config.help),
			Flag:        true,
		},
		{
			Names:       []string{"v"},
			Description: "Increase verbosity; may be specified more than once",
			Decoder:     writ.NewFlagAccumulator(&config.verbosity),
			Flag:        true,
			Plural:      true,
		},
		{
			Names:   []string{"t", "tolerance"},
			Decoder: writ.NewOptionDecoder(&config.tolerance),
		},
	}

	cmd.Help = writ.Help{
		Usage:  "Usage: explicit [OPTION]...",
		Header: "Explicit demonstrates explicit commands and options",
		Footer: "This method is flexible but more verbose than using writ.New()",
	}
	general := cmd.GroupOptions("help", "v")
	general.Header = "General Options:"
	cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, general)

	if runtime.GOOS == "darwin" {
		cmd.Options = append(cmd.Options, &writ.Option{
			Names:       []string{"use-quartz"},
			Description: "Use Quartz display on Mac",
			Decoder:     writ.NewFlagDecoder(&config.useQuartz),
			Flag:        true,
		})
		platform := cmd.GroupOptions("use-quartz")
		platform.Header = "Platform Options:"
		cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, platform)
	}

	// Decode the options
	_, _, err := cmd.Decode(os.Args[1:])
	if err != nil || config.help {
		cmd.ExitHelp(err)
	}
}
```

## Authors

Bob Ziuchkovski (@ziuchkovski)

## License (MIT)

Copyright (c) 2016 Bob Ziuchkovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


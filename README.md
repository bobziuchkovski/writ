[![Build Status](https://travis-ci.org/ziuchkovski/writ.svg?branch=master)](https://travis-ci.org/ziuchkovski/writ)
[![Coverage](http://gocover.io/_badge/github.com/ziuchkovski/writ?1)](http://gocover.io/github.com/ziuchkovski/writ)
[![Report Card](http://goreportcard.com/badge/ziuchkovski/writ)](http://goreportcard.com/report/ziuchkovski/writ)
[![GoDoc](https://godoc.org/github.com/ziuchkovski/writ?status.svg)](https://godoc.org/github.com/ziuchkovski/writ)

# Writ

## Overview

Writ is a flexible option parser with thorough test coverage.  It's meant to be simple and "just work".

Package writ implements option decoding with GNU getopt_long conventions. All long and short-form option variations are
supported: `--with-x`, `--name Sam`, `--day=Friday`, `-i FILE`, `-vvv`, etc.

## API Promise

Minor breaking changes may occur prior to the 1.0 release.  After 1.0 release, the API is guaranteed to remain backwards compatible.

## Basic Use

Please see the [godocs](https://godoc.org/github.com/ziuchkovski/writ) for additional information.

```go
package main

import (
    "fmt"
    "github.com/ziuchkovski/writ"
    "strings"
)

type Greeter struct {
    HelpFlag  bool   `flag:"help" description:"display this help message"`
    Verbosity int    `flag:"v, verbose" description:"display verbose output"`
    Name      string `option:"n, name" default:"Everyone" description:"the person to greet"`
}

// This example uses writ.New() to build a *writ.Command from the Greeter's
// struct fields.  The resulting *writ.Command decodes and updates the
// Greeter's fields in-place.  The Command.ExitHelp() method is used to
// display help content if --help is specified, or if invalid input
// arguments are received.
func main() {
    greeter := &Greeter{}
    cmd := writ.New("greeter", greeter)

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

### Explicit Commands and Options

Please see the [godocs](https://godoc.org/github.com/ziuchkovski/writ) for additional information.

```go
package main

import (
    "github.com/ziuchkovski/writ"
    "os"
    "runtime"
)

type Config struct {
    help      bool
    verbosity int
    useQuartz bool
}

// This example demonstrates explicit Command and Option creation,
// along with explicit option grouping.  It checks the host platform
// and dynamically adds a --use-quartz flag if the example is run on
// Mac OS.  The same result could be achieved by using writ.New() to
// construct a Command, and then adding the platform-specific option
// to the resulting Command directly.
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
    }

    general := cmd.GroupOptions("help", "v")
    general.Header = "General Options:"
    cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, general)

    // Dynamically add --with-quartz on Mac OS
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


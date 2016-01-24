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

Minor breaking changes may occur prior to the 1.0 release.  After the 1.0 release, the API is guaranteed to remain backwards compatible.

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
    help       bool
    verbosity  int
    bootloader string
}

// This example demonstrates explicit Command and Option creation,
// along with explicit option grouping.  It checks the host platform
// and dynamically adds a --bootloader option if the example is run on
// Linux.  The same result could be achieved by using writ.New() to
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

    // Dynamically add --bootloader on Linux
    if runtime.GOOS == "linux" {
        cmd.Options = append(cmd.Options, &writ.Option{
            Names:       []string{"bootloader"},
            Description: "Use the specified bootloader (grub, grub2, or lilo)",
            Decoder:     writ.NewOptionDecoder(&config.bootloader),
            Placeholder: "NAME",
        })
        platform := cmd.GroupOptions("bootloader")
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

### Subcommands

Please see the [godocs](https://godoc.org/github.com/ziuchkovski/writ) for additional information.

```go
package main

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
    if len(positional) == 0 {
        p.Last().ExitHelp(errors.New("ls requires one or more PATH arguments"))
    }
    // Listing operation omitted for brevity.  This would be a call to ioutil.ReadDir
    // followed by conditional formatting based on the r.LongFormat value.
}

// This example demonstrates subcommands in a busybox style.  There's no requirement
// that subcommands implement this particular Run() method, or any Run() method at
// at all.
func main() {
    gobox := &GoBox{}
    cmd := writ.New("gobox", gobox)
    cmd.Help.Usage = "Usage: gobox COMMAND [OPTION]... [ARG]..."
    cmd.Subcommand("ln").Help.Usage = "Usage: gobox ln [-s] OLD NEW"
    cmd.Subcommand("ls").Help.Usage = "Usage: gobox ls [-l] PATH [PATH]..."

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


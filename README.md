[![Build Status](https://travis-ci.org/bobziuchkovski/writ.svg?branch=master)](https://travis-ci.org/bobziuchkovski/writ)
[![Coverage](https://gocover.io/_badge/github.com/bobziuchkovski/writ?1)](https://gocover.io/github.com/bobziuchkovski/writ)
[![Report Card](http://goreportcard.com/badge/bobziuchkovski/writ)](http://goreportcard.com/report/bobziuchkovski/writ)
[![GoDoc](https://godoc.org/github.com/bobziuchkovski/writ?status.svg)](https://godoc.org/github.com/bobziuchkovski/writ)

# Writ

## Overview

Writ is a flexible option parser with thorough test coverage.  It's meant to be simple and "just work".  Applications
using writ look and behave similar to common GNU command-line applications, making them comfortable for end-users.

Writ implements option decoding with GNU getopt_long conventions. All long and short-form option variations are
supported: `--with-x`, `--name Sam`, `--day=Friday`, `-i FILE`, `-vvv`, etc.

Help output generation is supported using text/template.  The default template can be overriden with a custom template.

## API Promise

Minor breaking changes may occur prior to the 1.0 release.  After the 1.0 release, the API is guaranteed to remain backwards compatible.

## Basic Use

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/writ) for additional information.

This example uses writ.New() to build a command from the Greeter's struct fields.  The resulting *writ.Command decodes
and updates the Greeter's fields in-place.  The Command.ExitHelp() method is used to display help content if --help is
specified, or if invalid input arguments are received.

Source:

```go
package main

import (
    "fmt"
    "github.com/bobziuchkovski/writ"
    "strings"
)

type Greeter struct {
    HelpFlag  bool   `flag:"help" description:"Display this help message and exit"`
    Verbosity int    `flag:"v, verbose" description:"Display verbose output"`
    Name      string `option:"n, name" default:"Everyone" description:"The person or people to greet"`
}

func main() {
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
}
```

Help output:

```
Usage: greeter [OPTION]... MESSAGE
Greet users, displaying MESSAGE

Available Options:
  --help                    Display this help message and exit
  -v, --verbose             Display verbose output
  -n, --name=ARG            The person or people to greet
```


### Subcommands

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/writ) for additional information.

This example demonstrates subcommands in a busybox style.  There's no requirement that subcommands implement the Run()
method shown here.  It's just an example of how subcommands might be implemented.

Source:

```go
package main

import (
    "errors"
    "github.com/bobziuchkovski/writ"
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

func (g *GoBox) Run(p writ.Path, positional []string) {
    // The user didn't specify a subcommand.  Give them help.
    p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (l *Link) Run(p writ.Path, positional []string) {
    if l.HelpFlag {
        p.Last().ExitHelp(nil)
    }
    if len(positional) != 2 {
        p.Last().ExitHelp(errors.New("ln requires two arguments, OLD and NEW"))
    }
    // Link operation omitted for brevity.  This would be os.Link or os.Symlink
    // based on the l.Symlink value.
}

func (l *List) Run(p writ.Path, positional []string) {
    if l.HelpFlag {
        p.Last().ExitHelp(nil)
    }
    // Listing operation omitted for brevity.  This would be a call to ioutil.ReadDir
    // followed by conditional formatting based on the l.LongFormat value.
}

func main() {
    gobox := &GoBox{}
    cmd := writ.New("gobox", gobox)
    cmd.Help.Usage = "Usage: gobox COMMAND [OPTION]... [ARG]..."
    cmd.Subcommand("ln").Help.Usage = "Usage: gobox ln [-s] OLD NEW"
    cmd.Subcommand("ls").Help.Usage = "Usage: gobox ls [-l] [PATH]..."

    path, positional, err := cmd.Decode(os.Args[1:])
    if err != nil {
        // Using path.Last() here ensures the user sees relevant help for their
        // command selection
        path.Last().ExitHelp(err)
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

Help output, gobox:

```
Usage: gobox COMMAND [OPTION]... [ARG]...

Available Commands:
  ln                        Create a soft or hard link
  ls                        List directory contents
```

Help output, gobox ln:

```
Usage: gobox ln [-s] OLD NEW

Available Options:
  -h, --help                Display this message and exit
  -s                        Create a symlink instead of a hard link
```

Help output, gobox ls:

```
Usage: gobox ls [-l] [PATH]...

Available Options:
  -h, --help                Display this message and exit
  -l                        Use long-format output
```



### Explicit Commands and Options

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/writ) for additional information.

This example demonstrates explicit Command and Option creation, along with explicit option grouping.
It checks the host platform and dynamically adds a --bootloader option if the example is run on
Linux.  The same result could be achieved by using writ.New() to construct a Command, and then adding
the platform-specific option to the resulting Command directly.

Source:

```go
package main

import (
    "github.com/bobziuchkovski/writ"
    "os"
    "runtime"
)

type Config struct {
    help       bool
    verbosity  int
    bootloader string
}

func main() {
    config := &Config{}
    cmd := &writ.Command{Name: "explicit"}
    cmd.Help.Usage = "Usage: explicit [OPTION]... [ARG]..."
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

    // Note the explicit option grouping.  Using writ.New(), a single option group is
    // created for all options/flags that have descriptions.  Without writ.New(), we
    // need to create the OptionGroup(s) ourselves.
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

Help output, Linux:

```
General Options:
  -h, --help                Display this help text and exit
  -v                        Increase verbosity; may be specified more than once

Platform Options:
  --bootloader=NAME         Use the specified bootloader (grub, grub2, or lilo)
```

Help output, other platforms:

```
General Options:
  -h, --help                Display this help text and exit
  -v                        Increase verbosity; may be specified more than once
```

## Authors

Bob Ziuchkovski (@bobziuchkovski)

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


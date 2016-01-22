[![Build Status](https://travis-ci.org/ziuchkovski/writ.png?branch=master)](https://travis-ci.org/ziuchkovski/writ)
[![Report Card](http://goreportcard.com/badge/ziuchkovski/writ)](http://goreportcard.com/report/ziuchkovski/writ)
[![Coverage](http://gocover.io/_badge/github.com/ziuchkovski/writ?0)](http://gocover.io/github.com/ziuchkovski/writ)
[![GoDoc](https://godoc.org/github.com/ziuchkovski/writ?status.svg)](https://godoc.org/github.com/ziuchkovski/writ)

Writ
====

Overview
--------

Writ implements command line decoding according to [GNU getopt_long conventions](http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html).  All long and short-form option variations are supported: `--with-x`, `--name Sam`, `--day=Friday`, `-i FILE`, `-vvv`, etc.

Additionally, writ supports subcommands, customizable help output generation, and default values. However, writ is purely a decoder package. Command dispatch and execution are intentionally omitted.

Usage
-----

The following example is copied from writ's package documentation.  Please read the [godocs](https://godoc.org/github.com/ziuchkovski/writ) for additional information.


```golang
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
```


Authors
-------

Bob Ziuchkovski (@ziuchkovski)

License (MIT)
-------------

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


// Copyright (c) 2016 Bob Ziuchkovski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package writ

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var templateFuncs = map[string]interface{}{
	"formatCommand": formatCommand,
	"formatOption":  formatOption,
	"wrapText":      wrapText,
}

// The Help type is used for presentation purposes only, and does not affect
// argument parsing.
//
// The Command.ExitHelp() and Command.WriteHelp() methods execute the
// template assigned to the Template field, passing the Command as input.
// If the Template field is nil, the writ package's default template is used.
type Help struct {
	OptionGroups  []OptionGroup
	CommandGroups []CommandGroup

	// Optional
	Template *template.Template // Used to render output
	Usage    string             // Short message displayed at the top of output
	Header   string             // Displayed after Usage
	Footer   string             // Displayed at the end of output
}

// OptionGroup is used to customize help output.  It groups related Options
// for output.  When New() parses an input spec, it creates a single OptionGroup
// for all parsed options that have descriptions.
type OptionGroup struct {
	Options []*Option

	// Optional
	Name   string // Not displayed; for matching purposes within the template
	Header string // Displayed before the group
	Footer string // Displayed after the group
}

// CommandGroup is used to customize help output.  It groups related Commands
// for output.  When New() parses an input spec, it creates a single CommandGroup
// for all parsed commands that have descriptions.
type CommandGroup struct {
	Commands []*Command

	// Optional
	Name   string // Not displayed; for matching purposes within the template
	Header string // Displayed before the group
	Footer string // Displayed after the group
}

func formatOption(o *Option) string {
	var placeholder string
	if !o.Flag {
		placeholder = o.Placeholder
		if placeholder == "" {
			placeholder = "ARG"
		}
	}
	names := ""
	short := o.ShortNames()
	long := o.LongNames()
	for i, s := range short {
		names += "-" + s
		if (i < len(short)-1) || len(long) != 0 {
			names += ", "
		}
	}
	if len(long) == 0 && placeholder != "" {
		names += " " + placeholder
	}
	for i, l := range long {
		names += "--" + l
		if i < len(long)-1 {
			names += ", "
		} else if placeholder != "" {
			names += "=" + placeholder
		}
	}

	formatted := fmt.Sprintf("  %-24s  %s", names, o.Description)
	return wrapText(formatted, 80, 28)
}

func formatCommand(c *Command) string {
	formatted := fmt.Sprintf("  %-24s  %s", c.Name, c.Description)
	return wrapText(formatted, 80, 28)
}

// This is a pretty naiive implementation, but it's late and I'm tired
// TODO: cleanup and probably try to wrap on nearest space or punctuation
func wrapText(s string, width int, indent int) string {
	buf := bytes.NewBuffer(nil)
	runes := []rune(s)
	linelen, i := 0, 0
	for i < len(runes) {
		if runes[i] == '\n' {
			buf.WriteString("\n")
			if i < len(runes) {
				buf.WriteString(strings.Repeat(" ", indent))
				linelen = indent
			}
		} else if linelen == width {
			buf.WriteString("\n")
			if i < len(runes) {
				buf.WriteString(strings.Repeat(" ", indent))
				linelen = indent
			}
			buf.WriteRune(runes[i])
		} else {
			buf.WriteRune(runes[i])
		}
		i++
		linelen++
	}
	return buf.String()
}

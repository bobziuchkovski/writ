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
	"io/ioutil"
	"testing"
	"text/template"
)

var helpFormattingTests = []struct {
	Description string
	Spec        interface{}
	Rendered    string
}{
	{
		Description: "A single option",
		Spec: &struct {
			Flag bool `flag:"h, help" description:"Display this text and exit"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -h, --help                Display this text and exit
`,
	},

	{
		Description: "A couple options",
		Spec: &struct {
			Flag   bool `flag:"h" description:"Display this text and exit"`
			Option int  `option:"i, int" description:"An int option" placeholder:"INT"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -h                        Display this text and exit
  -i, --int=INT             An int option
`,
	},

	{
		Description: "Multiple long and short names for an option",
		Spec: &struct {
			Option int `option:"i, I, int, Int" description:"An int option" placeholder:"INT"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -i, -I, --int, --Int=INT  An int option
`,
	},

	{
		Description: "An option with short-form placeholder",
		Spec: &struct {
			Option int `option:"i" description:"An int option" placeholder:"INT"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -i INT                    An int option
`,
	},

	{
		Description: "A single command",
		Spec: &struct {
			Command struct{} `command:"command" description:"A command"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Commands:
  command                   A command
`,
	},

	{
		Description: "A single option and single command",
		Spec: &struct {
			Option  int      `option:"i" description:"An int option" placeholder:"INT"`
			Command struct{} `command:"command" description:"A command"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -i INT                    An int option

Available Commands:
  command                   A command
`,
	},

	{
		Description: "Command description wrapping",
		Spec: &struct {
			Command struct{} `command:"command" description:"A command with a reeeeeeeeeeeeeeeeeeeeeeeeeeeeeaaaaaaaaaallllllyyyyy loooooooooooooooonnnnnnngggggg description"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Commands:
  command                   A command with a reeeeeeeeeeeeeeeeeeeeeeeeeeeeeaaaaa
                            aaaaallllllyyyyy loooooooooooooooonnnnnnngggggg desc
                            ription
`,
	},

	{
		Description: "Command description wrapping with explicit newline in description",
		Spec: &struct {
			Command struct{} `command:"command" description:"A command with a\nnew line in the description"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Commands:
  command                   A command with a
                            new line in the description
`,
	},

	{
		Description: "Option description wrapping",
		Spec: &struct {
			Option int `option:"opt" description:"An option with a reeeeeeeeeeeeeeeeeeeeeeeeeeeeeaaaaaaaaaallllllyyyyy loooooooooooooooonnnnnnngggggg description"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  --opt=ARG                 An option with a reeeeeeeeeeeeeeeeeeeeeeeeeeeeeaaaaa
                            aaaaallllllyyyyy loooooooooooooooonnnnnnngggggg desc
                            ription
`,
	},

	{
		Description: "Option description wrapping with explicit newline in description",
		Spec: &struct {
			Option int `option:"opt" description:"An option with a\nnew line in the description"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  --opt=ARG                 An option with a
                            new line in the description
`,
	},

	{
		Description: "Hidden option",
		Spec: &struct {
			Hidden int  `option:"hidden"`
			Flag   bool `flag:"h, help" description:"Display this text and exit"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -h, --help                Display this text and exit
`,
	},

	{
		Description: "Hidden command",
		Spec: &struct {
			Command struct{} `command:"command" description:"A command"`
			Hidden  struct{} `command:"hidden"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Commands:
  command                   A command
`,
	},
}

func TestHelpFormatting(t *testing.T) {
	for _, test := range helpFormattingTests {
		cmd := New("test", test.Spec)
		buf := bytes.NewBuffer(nil)
		err := cmd.WriteHelp(buf)
		if err != nil {
			t.Errorf("Encountered unexpecting error running test.  Description: %s, Error: %s", test.Description, err)
			continue
		}
		if buf.String() != test.Rendered {
			t.Errorf("\nHelp output invalid.  Test Description: %s\n===Expected===\n%s\n\n===Received:===\n%s", test.Description, test.Rendered, buf.String())
			continue
		}
	}
}

func TestCustomHelpTemplate(t *testing.T) {
	templateText := "Custom content!"
	tpl := template.Must(template.New("Help").Parse(templateText))
	cmd := New("test", &struct{}{})
	cmd.Help.Template = tpl
	buf := bytes.NewBuffer(nil)
	err := cmd.WriteHelp(buf)
	if err != nil {
		t.Errorf("Encountered unexpecting error running custom template test.  Error: %s", err)
		return
	}
	if buf.String() != templateText {
		t.Errorf("Custom help output invalid.  Expected: %q, Received: %q", templateText, buf.String())
		return
	}
}

func TestInvalidHelpTemplate(t *testing.T) {
	templateText := "{{.Bogus}}"
	tpl := template.Must(template.New("Help").Parse(templateText))
	cmd := New("test", &struct{}{})
	cmd.Help.Template = tpl

	defer func() {
		r := recover()
		if r != nil {
			switch r.(type) {
			case commandError, optionError:
				// Intentionally blank
			default:
				panic(r)
			}
		}
	}()
	cmd.WriteHelp(ioutil.Discard)
	t.Errorf("Expected cmd.WriteHelp() to panic on invalid template, but this didn't happen")
}

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
	"testing"
	"text/template"
)

var helpFormattingTests = []struct {
	Description string
	Spec        interface{}
	Rendered    string
}{
	{
		Description: "Basic output 1",
		Spec: &struct {
			Flag bool `flag:"h, help" description:"Display this text and exit"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -h, --help                Display this text and exit
`,
	},

	{
		Description: "Basic output 2",
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
		Description: "Basic output 3",
		Spec: &struct {
			Option int `option:"i" description:"An int option" placeholder:"INT"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Options:
  -i INT                    An int option
`,
	},

	{
		Description: "Basic output 4",
		Spec: &struct {
			Command struct{} `command:"command" description:"A command"`
		}{},
		Rendered: `Usage: test [OPTION]... [ARG]...

Available Commands:
  command                   A command
`,
	},

	{
		Description: "Basic output 5",
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

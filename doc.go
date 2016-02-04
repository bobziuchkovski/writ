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

/*
Package writ implements a flexible option parser with thorough test coverage.
It's meant to be simple and "just work".  Applications using writ look and
behave similar to common GNU command-line applications, making them comfortable
for end-users.

Writ implements option decoding with GNU getopt_long conventions. All long and
short-form option variations are supported: --with-x, --name Sam, --day=Friday,
-i FILE, -vvv, etc.

Help output generation is supported using text/template.  The default template
can be overriden with a custom template.

Basics

Writ uses the Command and Option types to represent available options and
subcommands.  Input arguments are decoded with Command.Decode().

For convenience, the New() function can parse an input struct into a
Command with Options that represent the input struct's fields.  It uses
struct field tags to control the behavior.  The resulting Command's Decode()
method updates the struct's fields in-place when option arguments are decoded.

Alternatively, Commands and Options may be created directly.  All fields on
these types are exported.

Options

Options are specified via the "option" and "flag" struct tags.  Both represent
options, but fields marked "option" take arguments, whereas fields marked
"flag" do not.

Every Option must have an OptionDecoder.  Writ provides decoders for most
basic types, as well as some convenience types.  See the NewOptionDecoder()
function docs for details.

Commands

New() parses an input struct to build a top-level  Command.  Subcommands are
supported by using the "command" field tag.  Fields marked with "command" must
be of struct type, and are parsed the same way as top-level commands.

Help Output

Writ provides methods for generating help output.  Command.WriteHelp()
generates help content and writes to a given io.Writer.  Command.ExitHelp()
writes help content to stdout or stderr and terminates the program.

Writ uses a template to generate the help content.  The default template
mimics --help output for common GNU programs.  See the documentation of the
Help type for more details.

Field Tag Reference

The New() function recognizes the following combinations of field tags:

	Option Fields:
		- option (required): a comma-separated list of names for the option
		- description: the description to display for help output
		- placeholder: the placeholder value to use next to the option names (e.g. FILE)
		- default: the default value for the field
		- env: the name of an environment variable, the value of which is used as a default for the field

	Flag fields:
		- flag (required): a comma-separated list of names for the flag
		- description: the description to display for help output

	Command fields:
		- name (required): a name for the command
		- aliases: a comma-separated list of alias names for the command
		- description: the description to display for help output

If both "default" and "env" are specified for an option field, the environment
variable is consulted first.  If the environment variable is present and
decodes without error, that value is used.  Otherwise, the value for the
"default" tag is used.  Values specified via parsed arguments take precedence
over both types of defaults.
*/
package writ

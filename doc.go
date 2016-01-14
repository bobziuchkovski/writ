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
Package writ implements command line decoding according to GNU getopt_long
conventions: http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html
All long and short-form option variations are supported: --with-x, --name Sam,
--day=Friday, -i FILE, -vvv, etc.

Additionally, writ supports subcommands, customizable help output generation,
and default values.  However, writ is purely a decoder package.  Command
dispatch and execution are intentionally omitted.

Basics

Writ is modeled after the Go encoding/* packages.  The New() function
reads an input struct with relevant field tags and builds a Command that is
capable of decoding short and long-form command-line options into the input
struct's fields.

Options are specified via "flag" and "option" field tags, and subcommands are
specified via "command" field tags.  Additional field tags may be used to
specify command aliases, default values, and help content.

Once a Command is constructed, it's Decode() method is used to parse an input
[]string of arguments (e.g. os.Args[1:]).  The return values of Decode()
specify the parsed command path, remaining positional arguments, and any
parse/decode errors encountered.  The command path is a slice of []*Command
representing the invoked command hierarchy.  This can be used to differentiate
between user selection of top-level commands vs. subcommands.

Options

Options are specified via the "option" and "flag" struct tags.  Both are
considered options, but fields marked "option" take arguments, whereas
fields marked "flag" do not.

Field values are decoded via implementations of OptionDecoder.  Writ provides
OptionDecoder implementations for most basic field types along with several
convenience types.  See the NewOptionDecoder() documentation for a list of
supported types.  If a field implements the OptionDecoder interface, the
Decode() method from the field will be used.  Otherwise, writ will attempt
to generate an OptionDecoder for known field types.  If a field marked "flag"
or "option" is of an unsupported type and doesn't implement OptionDecoder,
then New() will panic when parsing the field.

Since flags take no arguments, they are valid only with bool, int, and
OptionDecoder fields.  Bool flags are true if the option is parsed, and int
flags maintain a count of the number of times the option is parsed.

Commands

New() parses an input struct with relevant field tags to build a top-level
Command.  Subcommands are supported by using the "command" field tag.  Fields
marked with "command" must be of struct type, and are parsed the same way as
top-level commands.  The resulting subcommands are then assigned to the
parent's Subcommand field.  The process repeats recursively.

Help Output

Writ provides helper methods for generating help output.  Command.WriteHelp()
generates help content and writes to a given io.Writer.  Command.ExitHelp()
goes a step further.  If the error argument is non-nil, it writes both the
help content and error message to Stderr, and then terminates the program with
an exit code of 1.  If the error argument is nil, it writes the help content to
os.Stdout and terminates the program with an exit code of 0.

The help output is designed to mimic the --help output for common GNU programs.
New() parses command and option descriptions from the "description" field tag.
Options may also specify the "placeholder" tag.

Help content is generated via the text/template package.  Writ provides a
default template that can be overridden on a per-Command basis.  See the
documentation of the Help type for more details.

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

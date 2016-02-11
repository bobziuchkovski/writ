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
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"
	"unicode"
)

type commandError struct {
	err error
}

func (e commandError) Error() string {
	return e.err.Error()
}

// panicCommand reports invalid use of the Command type
func panicCommand(format string, values ...interface{}) {
	e := commandError{fmt.Errorf(format, values...)}
	panic(e)
}

// Path represents a parsed Command list as returned by Command.Decode().
// It is used to differentiate between user selection of commands and
// subcommands.
type Path []*Command

// String returns the names of each command joined by spaces.
func (p Path) String() string {
	var parts []string
	for _, cmd := range p {
		parts = append(parts, cmd.Name)
	}
	return strings.Join(parts, " ")
}

// First returns the first command of the path.  This is the top-level/root command
// where Decode() was invoked.
func (p Path) First() *Command {
	return p[0]
}

// Last returns the last command of the path.  This is the user-selected command.
func (p Path) Last() *Command {
	return p[len(p)-1]
}

// findOption searches for the named option on the nearest ancestor command
func (p Path) findOption(name string) *Option {
	for i := len(p) - 1; i >= 0; i-- {
		o := p[i].Option(name)
		if o != nil {
			return o
		}
	}
	return nil
}

// New reads the input spec, searching for fields tagged with "option",
// "flag", or "command".  The field type and tags are used to construct
// a corresponding Command instance, which can be used to decode program
// arguments.  See the package overview documentation for details.
//
// NOTE: The spec value must be a pointer to a struct.
func New(name string, spec interface{}) *Command {
	cmd := parseCommandSpec(name, spec, nil)
	cmd.validate()
	return cmd
}

// Command specifies program options and subcommands.
//
// NOTE: If building a *Command directly without New(), the Help output
// will be empty by default.  Most applications will want to set the
// Help.Usage and Help.CommandGroups / Help.OptionGroups fields as
// appropriate.
type Command struct {
	// Required
	Name string

	// Optional
	Aliases     []string
	Options     []*Option
	Subcommands []*Command
	Help        Help
	Description string // Commands without descriptions are hidden
}

// String returns the command's name.
func (c *Command) String() string {
	return c.Name
}

// Decode parses the given arguments according to GNU getopt_long conventions.
// It matches Option arguments, both short and long-form, and decodes those
// arguments with the matched Option's Decoder field. If the Command has
// associated subcommands, the subcommand names are matched and extracted
// from the start of the positional arguments.
//
// To avoid ambiguity, subcommand matching terminates at the first unmatched
// positional argument.  Similarly, option names are matched against the
// command hierarchy as it exists at the point the option is encountered.  If
// command "first" has a subcommand "second", and "second" has an option
// "foo", then "first second --foo" is valid but "first --foo second" returns
// an error.  If the two commands, "first" and "second", both specify a "bar"
// option, then "first --bar second" decodes "bar" on "first", whereas
// "first second --bar" decodes "bar" on "second".
//
// As with GNU getopt_long, a bare "--" argument terminates argument parsing.
// All arguments after the first "--" argument are considered positional
// parameters.
func (c *Command) Decode(args []string) (path Path, positional []string, err error) {
	c.validate()
	c.setDefaults()
	return parseArgs(c, args)
}

// Subcommand locates subcommands on the method receiver.  It returns a match
// if any of the receiver's subcommands have a matching name or alias.  Otherwise
// it returns nil.
func (c *Command) Subcommand(name string) *Command {
	for _, sub := range c.Subcommands {
		if sub.Name == name {
			return sub
		}
		for _, a := range sub.Aliases {
			if a == name {
				return sub
			}
		}
	}
	return nil
}

// Option locates options on the method receiver.  It returns a match if any of
// the receiver's options have a matching name.  Otherwise it returns nil.  Options
// are searched only on the method receiver, not any of it's subcommands.
func (c *Command) Option(name string) *Option {
	for _, o := range c.Options {
		for _, n := range o.Names {
			if name == n {
				return o
			}
		}
	}
	return nil
}

// GroupOptions is used to build OptionGroups for help output.  It searches the
// method receiver for the named options and returns a corresponding OptionGroup.
// If any of the named options are not found, GroupOptions panics.
func (c *Command) GroupOptions(names ...string) OptionGroup {
	var group OptionGroup
	for _, n := range names {
		o := c.Option(n)
		if o == nil {
			panicCommand("Option not found: %s", n)
		}
		group.Options = append(group.Options, o)
	}
	return group
}

// GroupCommands is used to build CommandGroups for help output.  It searches the
// method receiver for the named subcommands and returns a corresponding CommandGroup.
// If any of the named subcommands are not found, GroupCommands panics.
func (c *Command) GroupCommands(names ...string) CommandGroup {
	var group CommandGroup
	for _, n := range names {
		c := c.Subcommand(n)
		if c == nil {
			panicCommand("Option not found: %s", n)
		}
		group.Commands = append(group.Commands, c)
	}
	return group
}

// WriteHelp renders help output to the given io.Writer.  Output is influenced
// by the Command's Help field.  See the Help type for details.
func (c *Command) WriteHelp(w io.Writer) error {
	var tmpl *template.Template
	if c.Help.Template != nil {
		tmpl = c.Help.Template
	} else {
		tmpl = defaultTemplate
	}

	buf := bytes.NewBuffer(nil)
	err := tmpl.Execute(buf, c)
	if err != nil {
		panicCommand("failed to render help: %s", err)
	}
	_, err = buf.WriteTo(w)
	return err
}

// ExitHelp writes help output and terminates the program.  If err is nil,
// the output is written to os.Stdout and the program terminates with a 0 exit
// code.  Otherwise, both the help output and error message are written to
// os.Stderr and the program terminates with a 1 exit code.
func (c *Command) ExitHelp(err error) {
	if err == nil {
		c.WriteHelp(os.Stdout)
		os.Exit(0)
	}
	c.WriteHelp(os.Stderr)
	fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
	os.Exit(1)
}

// validate command spec
func (c *Command) validate() {
	if c.Name == "" {
		panicCommand("Command name cannot be empty")
	}
	if strings.HasPrefix(c.Name, "-") {
		panicCommand("Command names cannot begin with '-' (command %s)", c.Name)
	}
	runes := []rune(c.Name)
	for _, r := range runes {
		if unicode.IsSpace(r) {
			panicCommand("Command names cannot have spaces (command %q)", c.Name)
		}
	}

	for _, a := range c.Aliases {
		if strings.HasPrefix(a, "-") {
			panicCommand("Command aliases cannot begin with '-' (command %s, alias %s)", c.Name, a)
		}
		runes := []rune(a)
		for _, r := range runes {
			if unicode.IsSpace(r) {
				panicCommand("Command aliases cannot have spaces (command %s, alias %q)", c.Name, a)
			}
		}
	}

	seen := make(map[string]bool)
	for _, sub := range c.Subcommands {
		sub.validate()
		subnames := append(sub.Aliases, sub.Name)
		for _, name := range subnames {
			_, present := seen[name]
			if present {
				panicCommand("command names must be unique (%s is specified multiple times)", name)
			}
			seen[name] = true
		}
	}

	seen = make(map[string]bool)
	for _, o := range c.Options {
		o.validate()
		for _, name := range o.Names {
			_, present := seen[name]
			if present {
				panicCommand("option names must be unique (%s is specified multiple times)", name)
			}
			seen[name] = true
		}
	}
}

func (c *Command) setDefaults() {
	for _, opt := range c.Options {
		defaulter, ok := opt.Decoder.(OptionDefaulter)
		if ok {
			defaulter.SetDefault()
		}
	}
	for _, sub := range c.Subcommands {
		sub.setDefaults()
	}
}

/*
 * Argument parsing
 */

func parseArgs(c *Command, args []string) (path Path, positional []string, err error) {
	path = Path{c}
	positional = make([]string, 0) // positional args should never be nil

	seen := make(map[*Option]bool)
	parseCmd, parseOpt := true, true
	for i := 0; i < len(args); i++ {
		a := args[i]
		if parseCmd {
			subcmd := path.Last().Subcommand(a)
			if subcmd != nil {
				path = append(path, subcmd)
				continue
			}
		}

		if parseOpt && strings.HasPrefix(a, "-") {
			if a == "-" {
				positional = append(positional, a)
				parseCmd = false
				continue
			}
			if a == "--" {
				parseOpt = false
				parseCmd = false
				continue
			}

			var opt *Option
			opt, args, err = processOption(path, args, i)
			if err != nil {
				return
			}
			_, present := seen[opt]
			if present && !opt.Plural {
				err = fmt.Errorf("option %q specified too many times", args[i])
				return
			}
			seen[opt] = true
			continue
		}

		// Unmatched positional arg
		parseCmd = false
		positional = append(positional, a)
	}
	return
}

func processOption(path Path, args []string, optidx int) (opt *Option, newargs []string, err error) {
	if strings.HasPrefix(args[optidx], "--") {
		return processLongOption(path, args, optidx)
	}
	return processShortOption(path, args, optidx)
}

func processLongOption(path Path, args []string, optidx int) (opt *Option, newargs []string, err error) {
	keyval := strings.SplitN(strings.TrimPrefix(args[optidx], "--"), "=", 2)
	name := keyval[0]
	newargs = args

	opt = path.findOption(name)
	if opt == nil {
		err = fmt.Errorf("option '--%s' is not recognized", name)
		return
	}
	if opt.Flag {
		if len(keyval) == 2 {
			err = fmt.Errorf("flag '--%s' does not accept an argument", name)
		} else {
			err = opt.Decoder.Decode("")
		}
	} else {
		if len(keyval) == 2 {
			err = opt.Decoder.Decode(keyval[1])
		} else {
			if len(args[optidx:]) < 2 {
				err = fmt.Errorf("option '--%s' requires an argument", name)
			} else {
				// Consume the next arg
				err = opt.Decoder.Decode(args[optidx+1])
				newargs = duplicateArgs(args)
				newargs = append(newargs[:optidx+1], newargs[optidx+2:]...)
			}
		}
	}
	return
}

func processShortOption(path Path, args []string, optidx int) (opt *Option, newargs []string, err error) {
	keyval := strings.SplitN(strings.TrimPrefix(args[optidx], "-"), "", 2)
	name := keyval[0]
	newargs = args

	opt = path.findOption(name)
	if opt == nil {
		err = fmt.Errorf("option '-%s' is not recognized", name)
		return
	}
	if opt.Flag {
		err = opt.Decoder.Decode("")
		if len(keyval) == 2 {
			// Short-form options are aggregated.  TODO: Cleanup
			// Rewrite current arg as -<name> and append remaining aggregate opts as a new arg after the current one
			newargs = duplicateArgs(args)
			newargs = append(newargs[:optidx+1], append([]string{"-" + keyval[1]}, newargs[optidx+1:]...)...)
			newargs[optidx] = "-" + name
		}
	} else {
		if len(keyval) == 2 {
			err = opt.Decoder.Decode(keyval[1])
		} else {
			if len(args[optidx:]) < 2 {
				err = fmt.Errorf("option '-%s' requires an argument", name)
			} else {
				// Consume the next arg
				err = opt.Decoder.Decode(args[optidx+1])
				newargs = duplicateArgs(args)
				newargs = append(newargs[:optidx+1], newargs[optidx+2:]...)
			}
		}
	}
	return
}

func duplicateArgs(args []string) []string {
	dupe := make([]string, len(args))
	for i := range args {
		dupe[i] = args[i]
	}
	return dupe
}

/*
 * Command spec parsing
 */

var (
	decoderPtr *OptionDecoder
	decoderT   = reflect.TypeOf(decoderPtr).Elem()

	aliasTag       = "alias"
	commandTag     = "command"
	defaultTag     = "default"
	descriptionTag = "description"
	envTag         = "env"
	flagTag        = "flag"
	optionTag      = "option"
	placeholderTag = "placeholder"
	invalidTags    = map[string][]string{
		commandTag: {defaultTag, envTag, flagTag, optionTag, placeholderTag},
		flagTag:    {aliasTag, commandTag, defaultTag, envTag, optionTag, placeholderTag},
		optionTag:  {aliasTag, commandTag, flagTag},
	}
)

func parseCommandSpec(name string, spec interface{}, path Path) *Command {
	rval := reflect.ValueOf(spec)
	if rval.Kind() != reflect.Ptr {
		panicCommand("command spec must be a pointer to struct type, not %s", rval.Kind())
	}
	if rval.Elem().Kind() != reflect.Struct {
		panicCommand("command spec must be a pointer to struct type, not %s", rval.Kind())
	}
	rval = rval.Elem()

	cmd := &Command{Name: name}
	path = append(path, cmd)

	for i := 0; i < rval.Type().NumField(); i++ {
		field := rval.Type().Field(i)
		fieldVal := rval.FieldByIndex(field.Index)
		if field.Tag.Get(commandTag) != "" {
			cmd.Subcommands = append(cmd.Subcommands, parseCommandField(field, fieldVal, path))
			continue
		}
		if field.Tag.Get(flagTag) != "" {
			cmd.Options = append(cmd.Options, parseFlagField(field, fieldVal))
			continue
		}
		if field.Tag.Get(optionTag) != "" {
			cmd.Options = append(cmd.Options, parseOptionField(field, fieldVal))
			continue
		}
	}

	var visibleOpts []*Option
	for _, opt := range cmd.Options {
		if opt.Description != "" {
			visibleOpts = append(visibleOpts, opt)
		}
	}
	if len(visibleOpts) > 0 {
		cmd.Help.OptionGroups = []OptionGroup{
			{Options: visibleOpts, Header: "Available Options:"},
		}
	}
	var visibleSubs []*Command
	for _, sub := range cmd.Subcommands {
		if sub.Description != "" {
			visibleSubs = append(visibleSubs, sub)
		}
	}
	if len(visibleSubs) > 0 {
		cmd.Help.CommandGroups = []CommandGroup{
			{Commands: visibleSubs, Header: "Available Commands:"},
		}
	}
	cmd.Help.Usage = fmt.Sprintf("Usage: %s [OPTION]... [ARG]...", path.String())
	return cmd
}

func parseCommandField(field reflect.StructField, fieldVal reflect.Value, path Path) *Command {
	checkTags(field, commandTag)
	checkExported(field, commandTag)

	names := parseCommaNames(field.Tag.Get(commandTag))
	if len(names) == 0 {
		panicCommand("commands must have a name (field %s)", field.Name)
	}
	if len(names) != 1 {
		panicCommand("commands must have a single name (field %s)", field.Name)
	}

	cmd := parseCommandSpec(names[0], fieldVal.Addr().Interface(), path)
	cmd.Aliases = parseCommaNames(field.Tag.Get(aliasTag))
	cmd.Description = field.Tag.Get(descriptionTag)
	cmd.validate()
	return cmd
}

func parseFlagField(field reflect.StructField, fieldVal reflect.Value) *Option {
	checkTags(field, flagTag)
	checkExported(field, flagTag)

	names := parseCommaNames(field.Tag.Get(flagTag))
	if len(names) == 0 {
		panicCommand("at least one flag name must be specified (field %s)", field.Name)
	}

	opt := &Option{
		Names:       names,
		Flag:        true,
		Description: field.Tag.Get(descriptionTag),
	}

	if field.Type.Implements(decoderT) {
		opt.Decoder = fieldVal.Interface().(OptionDecoder)
	} else if fieldVal.CanAddr() && reflect.PtrTo(field.Type).Implements(decoderT) {
		opt.Decoder = fieldVal.Addr().Interface().(OptionDecoder)
	} else {
		switch field.Type.Kind() {
		case reflect.Bool:
			opt.Decoder = NewFlagDecoder(fieldVal.Addr().Interface().(*bool))
		case reflect.Int:
			opt.Decoder = NewFlagAccumulator(fieldVal.Addr().Interface().(*int))
			opt.Plural = true
		default:
			panicCommand("field type not valid as a flag -- did you mean to use %q instead? (field %s)", "option", field.Name)
		}
	}

	opt.validate()
	return opt
}

func parseOptionField(field reflect.StructField, fieldVal reflect.Value) *Option {
	checkTags(field, optionTag)
	checkExported(field, optionTag)

	names := parseCommaNames(field.Tag.Get(optionTag))
	if len(names) == 0 {
		panicCommand("at least one option name must be specified (field %s)", field.Name)
	}

	opt := &Option{
		Names:       names,
		Description: field.Tag.Get(descriptionTag),
		Placeholder: field.Tag.Get(placeholderTag),
	}

	if field.Type.Implements(decoderT) {
		opt.Decoder = fieldVal.Interface().(OptionDecoder)
	} else if fieldVal.CanAddr() && reflect.PtrTo(field.Type).Implements(decoderT) {
		opt.Decoder = fieldVal.Addr().Interface().(OptionDecoder)
	} else {
		if fieldVal.Kind() == reflect.Bool {
			panicCommand("bool fields are not valid as options.  Use a %q tag instead (field %s)", "flag", field.Name)
		}
		if fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Map {
			opt.Plural = true
		}
		opt.Decoder = NewOptionDecoder(fieldVal.Addr().Interface())
	}

	defaultArg := field.Tag.Get(defaultTag)
	if defaultArg != "" {
		opt.Decoder = NewDefaulter(opt.Decoder, defaultArg)
	}
	envName := field.Tag.Get(envTag)
	if envName != "" {
		opt.Decoder = NewEnvDefaulter(opt.Decoder, envName)
	}

	opt.validate()
	return opt
}

func checkTags(field reflect.StructField, fieldType string) {
	badTags, present := invalidTags[fieldType]
	if !present {
		panic("BUG: fieldType not present in invalidTags map")
	}
	for _, t := range badTags {
		if field.Tag.Get(t) != "" {
			panicCommand("tag %s is not valid for %ss (field %s)", t, fieldType, field.Name)
		}
	}
}

func checkExported(field reflect.StructField, fieldType string) {
	if field.PkgPath != "" && !field.Anonymous {
		panicCommand("%ss must be exported (field %s)", fieldType, field.Name)
	}
}

func parseCommaNames(spec string) []string {
	isSep := func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	}
	return strings.FieldsFunc(spec, isSep)
}

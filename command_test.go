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
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func CompareField(structval interface{}, field string, value interface{}) (equal bool, fieldVal interface{}) {
	rval := reflect.ValueOf(structval)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	f := rval.FieldByName(field)
	equal = reflect.DeepEqual(f.Interface(), value)
	fieldVal = f.Interface()
	return
}

/*
 * Test command and option routing for multi-tier commands
 */

type topSpec struct {
	MidSpec midSpec `command:"mid" alias:"second, 2nd" description:"a mid-level command"`
	Top     int     `option:"t, topval" description:"an option on a top-level command"`
}

type midSpec struct {
	Mid        int        `option:"m, midval" description:"an option on a mid-level command"`
	BottomSpec bottomSpec `command:"bottom" alias:"third" description:"a bottom-level command"`
}

type bottomSpec struct {
	Bottom int `option:"b, bottomval" description:"an option on a bottom-level command"`
}

type commandFieldTest struct {
	Args       []string
	Valid      bool
	Path       string
	Positional []string
	Field      string
	Value      interface{}
	SkipReason string
}

var commandFieldTests = []commandFieldTest{
	// Path: top
	{Args: []string{}, Valid: true, Path: "top", Positional: []string{}},
	{Args: []string{"-"}, Valid: true, Path: "top", Positional: []string{"-"}},
	{Args: []string{"-", "mid"}, Valid: true, Path: "top", Positional: []string{"-", "mid"}},
	{Args: []string{"--"}, Valid: true, Path: "top", Positional: []string{}},
	{Args: []string{"--", "mid"}, Valid: true, Path: "top", Positional: []string{"mid"}},
	{Args: []string{"-t", "1"}, Valid: true, Path: "top", Positional: []string{}, Field: "Top", Value: 1},
	{Args: []string{"foo", "-t", "1"}, Valid: true, Path: "top", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "foo"}, Valid: true, Path: "top", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"foo", "bar"}, Valid: true, Path: "top", Positional: []string{"foo", "bar"}},
	{Args: []string{"foo", "-t", "1", "bar"}, Valid: true, Path: "top", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "foo", "bar"}, Valid: true, Path: "top", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"--", "mid"}, Valid: true, Path: "top", Positional: []string{"mid"}},
	{Args: []string{"-t", "1", "--", "mid"}, Valid: true, Path: "top", Positional: []string{"mid"}, Field: "Top", Value: 1},
	{Args: []string{"-", "-t", "1", "--", "mid"}, Valid: true, Path: "top", Positional: []string{"-", "mid"}, Field: "Top", Value: 1},
	{Args: []string{"--", "-t", "1", "mid"}, Valid: true, Path: "top", Positional: []string{"-t", "1", "mid"}, Field: "Top", Value: 0},
	{Args: []string{"--", "-t", "1", "-", "mid"}, Valid: true, Path: "top", Positional: []string{"-t", "1", "-", "mid"}, Field: "Top", Value: 0},
	{Args: []string{"bottom"}, Valid: true, Path: "top", Positional: []string{"bottom"}},
	{Args: []string{"third"}, Valid: true, Path: "top", Positional: []string{"third"}},
	{Args: []string{"bottom", "mid"}, Valid: true, Path: "top", Positional: []string{"bottom", "mid"}},
	{Args: []string{"bottom", "second"}, Valid: true, Path: "top", Positional: []string{"bottom", "second"}},
	{Args: []string{"bottom", "-", "second"}, Valid: true, Path: "top", Positional: []string{"bottom", "-", "second"}},
	{Args: []string{"-m", "2"}, Valid: false},
	{Args: []string{"--midval", "2"}, Valid: false},
	{Args: []string{"-b", "3"}, Valid: false},
	{Args: []string{"--bottomval", "3"}, Valid: false},
	{Args: []string{"--bogus", "4"}, Valid: false},
	{Args: []string{"--foo"}, Valid: false},
	{Args: []string{"--foo=bar"}, Valid: false},
	{Args: []string{"-f"}, Valid: false},
	{Args: []string{"-f", "bar"}, Valid: false},
	{Args: []string{"-fbar"}, Valid: false},

	// Path: top mid
	{Args: []string{"mid"}, Valid: true, Path: "top mid", Positional: []string{}},
	{Args: []string{"mid", "-"}, Valid: true, Path: "top mid", Positional: []string{"-"}},
	{Args: []string{"mid", "--"}, Valid: true, Path: "top mid", Positional: []string{}},
	{Args: []string{"mid", "-", "bottom"}, Valid: true, Path: "top mid", Positional: []string{"-", "bottom"}},
	{Args: []string{"mid", "--", "bottom"}, Valid: true, Path: "top mid", Positional: []string{"bottom"}},
	{Args: []string{"mid", "-t", "1"}, Valid: true, Path: "top mid", Positional: []string{}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-", "-t", "1"}, Valid: true, Path: "top mid", Positional: []string{"-"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "--", "-t", "1"}, Valid: true, Path: "top mid", Positional: []string{"-t", "1"}, Field: "Top", Value: 0},
	{Args: []string{"-t", "1", "mid"}, Valid: true, Path: "top mid", Positional: []string{}, Field: "Top", Value: 1},
	{Args: []string{"mid", "foo", "-t", "1"}, Valid: true, Path: "top mid", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "foo"}, Valid: true, Path: "top mid", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-t", "1", "foo"}, Valid: true, Path: "top mid", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "foo"}, Valid: true, Path: "top mid", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "foo", "-m", "2"}, Valid: true, Path: "top mid", Positional: []string{"foo"}, Field: "Mid", Value: 2},
	{Args: []string{"mid", "-m", "2", "foo"}, Valid: true, Path: "top mid", Positional: []string{"foo"}, Field: "Mid", Value: 2},
	{Args: []string{"mid", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}},
	{Args: []string{"second", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}},
	{Args: []string{"2nd", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}},
	{Args: []string{"mid", "foo", "-t", "1", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-t", "1", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "foo", "-m", "2", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Mid", Value: 2},
	{Args: []string{"-t", "1", "second", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"second", "foo", "-m", "2", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Mid", Value: 2},
	{Args: []string{"-t", "1", "2nd", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"2nd", "foo", "-m", "2", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Mid", Value: 2},
	{Args: []string{"mid", "-m", "2", "foo", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Mid", Value: 2},
	{Args: []string{"mid", "-m", "2", "foo", "-t", "1", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-m", "2", "foo", "-t", "1", "bar"}, Valid: true, Field: "Mid", Value: 2},
	{Args: []string{"mid", "-t", "1", "foo", "-m", "2", "bar"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-t", "1", "foo", "-m", "2", "bar"}, Valid: true, Field: "Mid", Value: 2},
	{Args: []string{"-t", "1", "mid", "foo", "bar", "-m", "2"}, Valid: true, Path: "top mid", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "foo", "bar", "-m", "2"}, Valid: true, Field: "Mid", Value: 2},
	{Args: []string{"mid", "-m", "2", "--"}, Valid: true, Path: "top mid", Positional: []string{}, Field: "Mid", Value: 2},
	{Args: []string{"mid", "--", "-m", "2"}, Valid: true, Path: "top mid", Positional: []string{"-m", "2"}, Field: "Mid", Value: 0},
	{Args: []string{"mid", "--", "bottom", "-b", "3"}, Valid: true, Path: "top mid", Positional: []string{"bottom", "-b", "3"}},
	{Args: []string{"mid", "--", "bottom", "-b", "3", "--"}, Valid: true, Path: "top mid", Positional: []string{"bottom", "-b", "3", "--"}},
	{Args: []string{"mid", "--", "bottom", "--", "-b", "3"}, Valid: true, Path: "top mid", Positional: []string{"bottom", "--", "-b", "3"}},
	{Args: []string{"-m", "2", "mid"}, Valid: false},
	{Args: []string{"--midval", "2", "mid"}, Valid: false},
	{Args: []string{"-m", "2", "mid", "foo"}, Valid: false},
	{Args: []string{"-b", "3", "mid"}, Valid: false},
	{Args: []string{"-b", "3", "mid", "foo"}, Valid: false},
	{Args: []string{"mid", "-b", "3"}, Valid: false},
	{Args: []string{"mid", "-b", "3", "foo"}, Valid: false},
	{Args: []string{"mid", "--bogus", "4"}, Valid: false},
	{Args: []string{"mid", "--foo"}, Valid: false},
	{Args: []string{"mid", "--foo=bar"}, Valid: false},
	{Args: []string{"mid", "-f"}, Valid: false},
	{Args: []string{"mid", "-f", "bar"}, Valid: false},
	{Args: []string{"mid", "-fbar"}, Valid: false},

	// Path: top mid bottom
	{Args: []string{"mid", "bottom"}, Valid: true, Path: "top mid bottom", Positional: []string{}},
	{Args: []string{"mid", "bottom", "-"}, Valid: true, Path: "top mid bottom", Positional: []string{"-"}},
	{Args: []string{"mid", "bottom", "--"}, Valid: true, Path: "top mid bottom", Positional: []string{}},
	{Args: []string{"mid", "bottom", "-t", "1"}, Valid: true, Path: "top mid bottom", Positional: []string{}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-t", "1", "bottom"}, Valid: true, Path: "top mid bottom", Positional: []string{}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "bottom"}, Valid: true, Path: "top mid bottom", Positional: []string{}, Field: "Top", Value: 1},
	{Args: []string{"mid", "bottom", "foo", "-t", "1"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "-t", "1", "bottom", "foo"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "bottom", "foo"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "bottom", "foo", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "bottom", "-b", "3", "foo"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "bottom", "foo", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}},
	{Args: []string{"mid", "third", "-b", "3", "foo"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo"}, Field: "Bottom", Value: 3},
	{Args: []string{"2nd", "third", "foo", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}},
	{Args: []string{"-t", "1", "mid", "bottom", "foo", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Top", Value: 1},
	{Args: []string{"mid", "bottom", "foo", "-b", "3", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "bottom", "-b", "3", "foo", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "-t", "1", "bottom", "-b", "3", "foo", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "-t", "1", "bottom", "-b", "3", "foo", "bar"}, Valid: true, Field: "Top", Value: 1},
	{Args: []string{"mid", "-m", "2", "bottom", "foo", "-b", "3", "bar"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "-m", "2", "bottom", "foo", "-b", "3", "bar"}, Valid: true, Field: "Mid", Value: 2},
	{Args: []string{"-t", "1", "mid", "bottom", "foo", "bar", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"-t", "1", "mid", "bottom", "foo", "bar", "-b", "3"}, Valid: true, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "2nd", "bottom", "foo", "bar", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"-t", "1", "2nd", "bottom", "foo", "bar", "-b", "3"}, Valid: true, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "mid", "third", "foo", "bar", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"-t", "1", "mid", "third", "foo", "bar", "-b", "3"}, Valid: true, Field: "Top", Value: 1},
	{Args: []string{"-t", "1", "second", "third", "foo", "bar", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"foo", "bar"}, Field: "Bottom", Value: 3},
	{Args: []string{"-t", "1", "second", "third", "foo", "bar", "-b", "3"}, Valid: true, Field: "Top", Value: 1},
	{Args: []string{"mid", "bottom", "-b", "3", "--"}, Valid: true, Path: "top mid bottom", Positional: []string{}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "bottom", "-", "-b", "3", "--"}, Valid: true, Path: "top mid bottom", Positional: []string{"-"}, Field: "Bottom", Value: 3},
	{Args: []string{"mid", "bottom", "--", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"-b", "3"}, Field: "Bottom", Value: 0},
	{Args: []string{"mid", "bottom", "-", "--", "-b", "3"}, Valid: true, Path: "top mid bottom", Positional: []string{"-", "-b", "3"}, Field: "Bottom", Value: 0},
	{Args: []string{"mid", "-b", "3", "bottom"}, Valid: false},
	{Args: []string{"bottom", "-b", "3"}, Valid: false},
	{Args: []string{"-b", "3", "bottom"}, Valid: false},
	{Args: []string{"-b", "3", "mid", "bottom"}, Valid: false},
}

func TestCommandFields(t *testing.T) {
	for _, test := range commandFieldTests {
		spec := &topSpec{}
		runCommandFieldTest(t, spec, test)
	}
}

func runCommandFieldTest(t *testing.T, spec *topSpec, test commandFieldTest) {
	if test.SkipReason != "" {
		t.Logf("Test skipped. Args: %q, Field: %s, Reason: %s", test.Args, test.Field, test.SkipReason)
		return
	}

	cmd := New("top", spec)
	path, positional, err := cmd.Decode(test.Args)
	values := map[string]interface{}{
		"Top":    spec.Top,
		"Mid":    spec.MidSpec.Mid,
		"Bottom": spec.MidSpec.BottomSpec.Bottom,
	}
	if !test.Valid {
		if err == nil {
			t.Errorf("Expected error but none received. Args: %q", test.Args)
		}
		return
	}
	if err != nil {
		t.Errorf("Received unexpected error. Field: %s, Args: %q, Error: %s", test.Field, test.Args, err)
		return
	}
	if test.Positional != nil && !reflect.DeepEqual(positional, test.Positional) {
		t.Errorf("Positional args are incorrect. Args: %q, Expected: %s, Received: %s", test.Args, test.Positional, positional)
		return
	}
	if test.Field != "" && !reflect.DeepEqual(values[test.Field], test.Value) {
		t.Errorf("Decoded value is incorrect. Field: %s, Args: %q, Expected: %#v, Received: %#v", test.Field, test.Args, test.Value, values[test.Field])
		return
	}
	if path.First() != cmd {
		t.Errorf("Expected first command in path to be top-level command, but got %s instead.", path.First().Name)
		return
	}
	if test.Path != "" && path.String() != test.Path {
		t.Errorf("Command path is incorrect. Args: %q, Expected: %s, Received: %s", test.Args, test.Path, path)
		return
	}
}

func TestCommandString(t *testing.T) {
	cmd := New("top", &topSpec{})
	if cmd.String() != "top" {
		t.Errorf("Invalid Command.String() value.  Expected: %q, received: %q", "top", cmd.String())
	}
	if cmd.Subcommand("mid").String() != "mid" {
		t.Errorf("Invalid Command.String() value.  Expected: %q, received: %q", "mid", cmd.Subcommand("mid").String())
	}
	if cmd.Subcommand("mid").Subcommand("bottom").String() != "bottom" {
		t.Errorf("Invalid Command.String() value.  Expected: %q, received: %q", "bottom", cmd.Subcommand("mid").Subcommand("bottom").String())
	}
}

/*
 * Test parsing of description metadata
 */

func TestSpecDescriptions(t *testing.T) {
	type Spec struct {
		Flag    bool     `flag:"flag" description:"a flag"`
		Option  int      `option:"option" description:"an option"`
		Command struct{} `command:"command" description:"a command"`
	}
	cmd := New("test", &Spec{})
	if cmd.Option("flag").Description != "a flag" {
		t.Errorf("Flag description is incorrect. Expected: %q, Received: %q", "a flag", cmd.Option("flag").Description)
	}
	if cmd.Option("option").Description != "an option" {
		t.Errorf("Option description is incorrect. Expected: %q, Received: %q", "an option", cmd.Option("option").Description)
	}
	if cmd.Subcommand("command").Description != "a command" {
		t.Errorf("Command description is incorrect. Expected: %q, Received: %q", "a command", cmd.Subcommand("command").Description)
	}
}

/*
 * Test parsing of placeholder metadata
 */

func TestSpecPlaceholders(t *testing.T) {
	type Spec struct {
		Option int `option:"option" description:"an option" placeholder:"VALUE"`
	}
	cmd := New("test", &Spec{})
	if cmd.Option("option").Placeholder != "VALUE" {
		t.Errorf("Option placeholder is incorrect. Expected: %q, Received: %q", "VALUE", cmd.Option("option").Placeholder)
	}
}

/*
 * Test default values on fields
 */

type defaultFieldSpec struct {
	Default        int `option:"d" description:"An int field with a default" default:"42"`
	EnvDefault     int `option:"e" description:"An int field with an environment default" env:"ENV_DEFAULT"`
	StackedDefault int `option:"s" description:"An int dield with both a default and environment default" default:"84" env:"STACKED_DEFAULT"`
}

type defaultFieldTest struct {
	Args       []string
	Valid      bool
	Field      string
	Value      interface{}
	EnvKey     string
	EnvValue   string
	SkipReason string
}

var defaultFieldTests = []defaultFieldTest{
	// Field with a default value
	{Args: []string{""}, Valid: true, Field: "Default", Value: 42},
	{Args: []string{"-d", "2"}, Valid: true, Field: "Default", Value: 2},
	{Args: []string{"-d", "foo"}, Valid: false},

	// Field with an environment default
	{Args: []string{""}, Valid: true, Field: "EnvDefault", Value: 0},
	{Args: []string{""}, Valid: true, EnvKey: "ENV_DEFAULT", EnvValue: "2", Field: "EnvDefault", Value: 2},
	{Args: []string{""}, Valid: true, EnvKey: "ENV_DEFAULT", EnvValue: "foo", Field: "EnvDefault", Value: 0},
	{Args: []string{"-e", "4"}, Valid: true, EnvKey: "ENV_DEFAULT", EnvValue: "2", Field: "EnvDefault", Value: 4},
	{Args: []string{"-e", "4"}, Valid: true, EnvKey: "ENV_DEFAULT", EnvValue: "foo", Field: "EnvDefault", Value: 4},
	{Args: []string{"-e", "foo"}, Valid: false, EnvKey: "ENV_DEFAULT", EnvValue: "2"},

	// Field with both a default value and an environment default
	{Args: []string{""}, Valid: true, Field: "StackedDefault", Value: 84},
	{Args: []string{""}, Valid: true, EnvKey: "STACKED_DEFAULT", EnvValue: "2", Field: "StackedDefault", Value: 2},
	{Args: []string{""}, Valid: true, EnvKey: "STACKED_DEFAULT", EnvValue: "foo", Field: "StackedDefault", Value: 84},
	{Args: []string{"-s", "4"}, Valid: true, EnvKey: "STACKED_DEFAULT", EnvValue: "2", Field: "StackedDefault", Value: 4},
	{Args: []string{"-s", "4"}, Valid: true, EnvKey: "STACKED_DEFAULT", EnvValue: "foo", Field: "StackedDefault", Value: 4},
	{Args: []string{"-s", "foo"}, Valid: false, EnvKey: "STACKED_DEFAULT", EnvValue: "foo"},
	{Args: []string{"-s", "foo"}, Valid: false},
}

func TestDefaultFields(t *testing.T) {
	for _, test := range defaultFieldTests {
		spec := &defaultFieldSpec{}
		runDefaultFieldTest(t, spec, test)
	}
}

func runDefaultFieldTest(t *testing.T, spec interface{}, test defaultFieldTest) {
	if test.SkipReason != "" {
		t.Logf("Test skipped. Args: %q, Field: %s, Reason: %s", test.Args, test.Field, test.SkipReason)
		return
	}

	if test.EnvKey != "" {
		realval := os.Getenv(test.EnvKey)
		defer (func() { os.Setenv(test.EnvKey, realval) })()
		os.Setenv(test.EnvKey, test.EnvValue)
	}
	cmd := New("test", spec)
	_, _, err := cmd.Decode(test.Args)

	if !test.Valid {
		if err == nil {
			t.Errorf("Expected error but none received. Args: %q", test.Args)
		}
		return
	}
	if err != nil {
		t.Errorf("Received unexpected error. Field: %s, Args: %q, Error: %s", test.Field, test.Args, err)
		return
	}
	equal, fieldval := CompareField(spec, test.Field, test.Value)
	if !equal {
		t.Errorf("Decoded value is incorrect. Field: %s, Args: %q, Expected: %#v, Received: %#v", test.Field, test.Args, test.Value, fieldval)
		return
	}
}

/*
 * Generic field test helpers
 */

type fieldTest struct {
	Args       []string
	Valid      bool
	Field      string
	Value      interface{}
	SkipReason string
}

func runFieldTest(t *testing.T, spec interface{}, test fieldTest) {
	if test.SkipReason != "" {
		t.Logf("Test skipped. Args: %q, Field: %s, Reason: %s", test.Args, test.Field, test.SkipReason)
		return
	}

	cmd := New("test", spec)
	_, _, err := cmd.Decode(test.Args)
	if !test.Valid {
		if err == nil {
			t.Errorf("Expected error but none received. Args: %q", test.Args)
		}
		return
	}
	if err != nil {
		t.Errorf("Received unexpected error. Field: %s, Args: %q, Error: %s", test.Field, test.Args, err)
		return
	}
	equal, fieldval := CompareField(spec, test.Field, test.Value)
	if !equal {
		t.Errorf("Decoded value is incorrect. Field: %s, Args: %q, Expected: %#v, Received: %#v", test.Field, test.Args, test.Value, fieldval)
		return
	}
}

/*
 * Test option name variations
 */

type optNameSpec struct {
	Bool        bool    `flag:" b, bool" description:"A bool flag"`
	Accumulator int     `flag:"a,acc,A, accum" description:"An accumulator field"`
	Int         int     `option:"int,  I" description:"An int field"`
	Float       float32 `option:" float,F, FloaT,  f " description:"A float field"`
}

var optNameTests = []fieldTest{
	// Bool Flag
	{Args: []string{"-b"}, Valid: true, Field: "Bool", Value: true},
	{Args: []string{"--bool"}, Valid: true, Field: "Bool", Value: true},

	// Accumulator Flag
	{Args: []string{"-A", "--accum", "-a", "--acc", "-Aa", "-aA"}, Valid: true, Field: "Accumulator", Value: 8},

	// Int Option
	{Args: []string{"-I2"}, Valid: true, Field: "Int", Value: 2},
	{Args: []string{"-I", "2"}, Valid: true, Field: "Int", Value: 2},
	{Args: []string{"--int", "2"}, Valid: true, Field: "Int", Value: 2},
	{Args: []string{"--int=2"}, Valid: true, Field: "Int", Value: 2},

	// Float Option
	{Args: []string{"-F2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"-F2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"-F", "2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"-F", "2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"-f2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"-f2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"-f", "2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"-f", "2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"--FloaT", "2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"--FloaT", "2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"--FloaT=2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"--FloaT=2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"--float", "2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"--float", "2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
	{Args: []string{"--float=2"}, Valid: true, Field: "Float", Value: float32(2.0)},
	{Args: []string{"--float=2.5"}, Valid: true, Field: "Float", Value: float32(2.5)},
}

func TestOptionNames(t *testing.T) {
	for _, test := range optNameTests {
		spec := &optNameSpec{}
		runFieldTest(t, spec, test)
	}
}

/*
 * Test flag fields
 */

type flagFieldSpec struct {
	Bool        bool `flag:"b, bool" description:"A bool flag"`
	Accumulator int  `flag:"a, acc" description:"An accumulator flag"`
}

var flagTests = []fieldTest{
	// Bool flag
	{Args: []string{}, Valid: true, Field: "Bool", Value: false},
	{Args: []string{"-b"}, Valid: true, Field: "Bool", Value: true},
	{Args: []string{"--bool"}, Valid: true, Field: "Bool", Value: true},
	{Args: []string{"-b", "-b"}, Valid: false},
	{Args: []string{"-b2"}, Valid: false},
	{Args: []string{"--bool=2"}, Valid: false},

	// Accumulator flag
	{Args: []string{}, Valid: true, Field: "Accumulator", Value: 0},
	{Args: []string{"-a"}, Valid: true, Field: "Accumulator", Value: 1},
	{Args: []string{"-a", "-a"}, Valid: true, Field: "Accumulator", Value: 2},
	{Args: []string{"-aaa"}, Valid: true, Field: "Accumulator", Value: 3},
	{Args: []string{"--acc", "-a"}, Valid: true, Field: "Accumulator", Value: 2},
	{Args: []string{"-a", "--acc", "-aa"}, Valid: true, Field: "Accumulator", Value: 4},
	{Args: []string{"-a3"}, Valid: false},
	{Args: []string{"--acc=3"}, Valid: false},
}

func TestFlagFields(t *testing.T) {
	for _, test := range flagTests {
		spec := &flagFieldSpec{}
		runFieldTest(t, spec, test)
	}
}

/*
 * Test map and slice field types
 */

type mapSliceFieldSpec struct {
	StringSlice []string          `option:"s" description:"A string slice option" placeholder:"STRINGSLICE"`
	StringMap   map[string]string `option:"m" description:"A map of strings option" placeholder:"KEY=VALUE"`
}

var mapSliceFieldTests = []fieldTest{
	// String Slice
	{Args: []string{"-s", "1", "-s", "-1", "-s", "+1"}, Valid: true, Field: "StringSlice", Value: []string{"1", "-1", "+1"}},
	{Args: []string{"-s", " a b", "-s", "\n", "-s", "\t"}, Valid: true, Field: "StringSlice", Value: []string{" a b", "\n", "\t"}},
	{Args: []string{"-s", "日本", "-s", "-日本", "-s", "--日本"}, Valid: true, Field: "StringSlice", Value: []string{"日本", "-日本", "--日本"}},
	{Args: []string{"-s", "1"}, Valid: true, Field: "StringSlice", Value: []string{"1"}},
	{Args: []string{"-s", "-1"}, Valid: true, Field: "StringSlice", Value: []string{"-1"}},
	{Args: []string{"-s", "+1"}, Valid: true, Field: "StringSlice", Value: []string{"+1"}},
	{Args: []string{"-s", "1.0"}, Valid: true, Field: "StringSlice", Value: []string{"1.0"}},
	{Args: []string{"-s", "0x01"}, Valid: true, Field: "StringSlice", Value: []string{"0x01"}},
	{Args: []string{"-s", "-"}, Valid: true, Field: "StringSlice", Value: []string{"-"}},
	{Args: []string{"-s", "-a"}, Valid: true, Field: "StringSlice", Value: []string{"-a"}},
	{Args: []string{"-s", "--"}, Valid: true, Field: "StringSlice", Value: []string{"--"}},
	{Args: []string{"-s", "--a"}, Valid: true, Field: "StringSlice", Value: []string{"--a"}},
	{Args: []string{"-s", ""}, Valid: true, Field: "StringSlice", Value: []string{""}},
	{Args: []string{"-s", " "}, Valid: true, Field: "StringSlice", Value: []string{" "}},
	{Args: []string{"-s", " a"}, Valid: true, Field: "StringSlice", Value: []string{" a"}},
	{Args: []string{"-s", "a "}, Valid: true, Field: "StringSlice", Value: []string{"a "}},
	{Args: []string{"-s", "a b "}, Valid: true, Field: "StringSlice", Value: []string{"a b "}},
	{Args: []string{"-s", " a b"}, Valid: true, Field: "StringSlice", Value: []string{" a b"}},
	{Args: []string{"-s", "\n"}, Valid: true, Field: "StringSlice", Value: []string{"\n"}},
	{Args: []string{"-s", "\t"}, Valid: true, Field: "StringSlice", Value: []string{"\t"}},
	{Args: []string{"-s", "日本"}, Valid: true, Field: "StringSlice", Value: []string{"日本"}},
	{Args: []string{"-s", "-日本"}, Valid: true, Field: "StringSlice", Value: []string{"-日本"}},
	{Args: []string{"-s", "--日本"}, Valid: true, Field: "StringSlice", Value: []string{"--日本"}},
	{Args: []string{"-s", " 日本"}, Valid: true, Field: "StringSlice", Value: []string{" 日本"}},
	{Args: []string{"-s", "日本 "}, Valid: true, Field: "StringSlice", Value: []string{"日本 "}},
	{Args: []string{"-s", "日 本"}, Valid: true, Field: "StringSlice", Value: []string{"日 本"}},
	{Args: []string{"-s", "A relatively long string to make sure we aren't doing any silly truncation anywhere, since that would be bad..."}, Valid: true, Field: "StringSlice", Value: []string{"A relatively long string to make sure we aren't doing any silly truncation anywhere, since that would be bad..."}},
	{Args: []string{"-s"}, Valid: false},

	// String Map
	{Args: []string{"-m", "a=b"}, Valid: true, Field: "StringMap", Value: map[string]string{"a": "b"}},
	{Args: []string{"-m", "a=b=c"}, Valid: true, Field: "StringMap", Value: map[string]string{"a": "b=c"}},
	{Args: []string{"-m", "a=b "}, Valid: true, Field: "StringMap", Value: map[string]string{"a": "b "}},
	{Args: []string{"-m", "a= b"}, Valid: true, Field: "StringMap", Value: map[string]string{"a": " b"}},
	{Args: []string{"-m", "a =b"}, Valid: true, Field: "StringMap", Value: map[string]string{"a ": "b"}},
	{Args: []string{"-m", " a=b"}, Valid: true, Field: "StringMap", Value: map[string]string{" a": "b"}},
	{Args: []string{"-m", " a=b "}, Valid: true, Field: "StringMap", Value: map[string]string{" a": "b "}},
	{Args: []string{"-m", "a = b "}, Valid: true, Field: "StringMap", Value: map[string]string{"a ": " b "}},
	{Args: []string{"-m", " a = b "}, Valid: true, Field: "StringMap", Value: map[string]string{" a ": " b "}},
	{Args: []string{"-m", "a=b", "-m", "a=c"}, Valid: true, Field: "StringMap", Value: map[string]string{"a": "c"}},
	{Args: []string{"-m", "a=b", "-m", "c=d"}, Valid: true, Field: "StringMap", Value: map[string]string{"a": "b", "c": "d"}},
	{Args: []string{"-m", "日=本", "-m", "-日=本", "-m", "--日=--本"}, Valid: true, Field: "StringMap", Value: map[string]string{"日": "本", "-日": "本", "--日": "--本"}},
	{Args: []string{"-m"}, Valid: false},
}

func TestMapSliceFields(t *testing.T) {
	for _, test := range mapSliceFieldTests {
		spec := &mapSliceFieldSpec{}
		runFieldTest(t, spec, test)
	}
}

/*
 * Test io field types
 */

const ioTestText = "test IO"

type ioFieldSpec struct {
	Reader      io.Reader      `option:"reader" description:"An io.Reader input option"`
	ReadCloser  io.ReadCloser  `option:"readcloser" description:"An io.ReadCloser input option"`
	Writer      io.Writer      `option:"writer" description:"An io.Writer output option"`
	WriteCloser io.WriteCloser `option:"writecloser" description:"An io.WriteCloser output option"`
}

func (r *ioFieldSpec) PerformIO() error {
	input := []io.Reader{r.Reader, r.ReadCloser}
	for _, in := range input {
		if in != nil {
			bytes, err := ioutil.ReadAll(in)
			if err != nil {
				return err
			}
			if string(bytes) != ioTestText {
				return fmt.Errorf("Expected to read %q. Read %q instead.", ioTestText, string(bytes))
			}
			closer, ok := in.(io.Closer)
			if ok {
				err = closer.Close()
				if err != nil {
					return err
				}
			}
		}
	}
	output := []io.Writer{r.Writer, r.WriteCloser}
	for _, out := range output {
		if out != nil {
			_, err := io.WriteString(out, ioTestText)
			if err != nil {
				return err
			}
			closer, ok := out.(io.Closer)
			if ok {
				err = closer.Close()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type ioFieldTest struct {
	Args       []string
	Valid      bool
	Field      string
	InFiles    []string
	OutFiles   []string
	SkipReason string
}

var ioFieldTests = []ioFieldTest{
	// No-op
	{Args: []string{}, Valid: true, InFiles: []string{}, OutFiles: []string{}},

	// io.Reader
	{Args: []string{"--reader", "-"}, Valid: true, Field: "Reader", InFiles: []string{"stdin"}, OutFiles: []string{}},
	{Args: []string{"--reader", "infile"}, Valid: true, Field: "Reader", InFiles: []string{"infile"}, OutFiles: []string{}},
	{Args: []string{"--reader", "bogus/infile"}, Valid: false},
	{Args: []string{"--reader", ""}, Valid: false},
	{Args: []string{"--reader", "infile1", "--reader", "intput2"}, Valid: false},
	{Args: []string{"--reader", "-", "--reader", "intput"}, Valid: false},
	{Args: []string{"--reader", "infile", "--reader", "-"}, Valid: false},
	{Args: []string{"--reader"}, Valid: false},

	// io.ReadCloser
	{Args: []string{"--readcloser", "-"}, Valid: true, Field: "ReadCloser", InFiles: []string{"stdin"}, OutFiles: []string{}},
	{Args: []string{"--readcloser", "infile"}, Valid: true, Field: "ReadCloser", InFiles: []string{"infile"}, OutFiles: []string{}},
	{Args: []string{"--readcloser", "bogus/infile"}, Valid: false},
	{Args: []string{"--readcloser", ""}, Valid: false},
	{Args: []string{"--readcloser", "infile1", "--readcloser", "intput2"}, Valid: false},
	{Args: []string{"--readcloser", "-", "--readcloser", "intput"}, Valid: false},
	{Args: []string{"--readcloser", "infile", "--readcloser", "-"}, Valid: false},
	{Args: []string{"--readcloser"}, Valid: false},

	// io.Writer
	{Args: []string{"--writer", "-"}, Valid: true, Field: "Writer", InFiles: []string{}, OutFiles: []string{"stdout"}},
	{Args: []string{"--writer", "outfile"}, Valid: true, Field: "Writer", InFiles: []string{}, OutFiles: []string{"outfile"}},
	{Args: []string{"--writer", ""}, Valid: false},
	{Args: []string{"--writer", "bogus/outfile"}, Valid: false},
	{Args: []string{"--writer", "outfile1", "--writer", "outfile2"}, Valid: false},
	{Args: []string{"--writer", "-", "--writer", "outfile"}, Valid: false},
	{Args: []string{"--writer", "outfile", "--writer", "-"}, Valid: false},
	{Args: []string{"--writer"}, Valid: false},

	// io.WriteCloser
	{Args: []string{"--writecloser", "-"}, Valid: true, Field: "WriteCloser", InFiles: []string{}, OutFiles: []string{"stdout"}},
	{Args: []string{"--writecloser", "outfile"}, Valid: true, Field: "WriteCloser", InFiles: []string{}, OutFiles: []string{"outfile"}},
	{Args: []string{"--writecloser", ""}, Valid: false},
	{Args: []string{"--writecloser", "bogus/outfile"}, Valid: false},
	{Args: []string{"--writecloser", "outfile1", "--writecloser", "outfile2"}, Valid: false},
	{Args: []string{"--writecloser", "-", "--writecloser", "outfile"}, Valid: false},
	{Args: []string{"--writecloser", "outfile", "--writecloser", "-"}, Valid: false},
	{Args: []string{"--writecloser"}, Valid: false},
}

func TestIOFields(t *testing.T) {
	for _, test := range ioFieldTests {
		spec := &ioFieldSpec{}
		runIOFieldTest(t, spec, test)
	}
}

func runIOFieldTest(t *testing.T, spec *ioFieldSpec, test ioFieldTest) {
	if test.SkipReason != "" {
		t.Logf("Test skipped. Args: %q, Field: %s, Reason: %s", test.Args, test.Field, test.SkipReason)
		return
	}

	realin, realout := os.Stdin, os.Stdout
	defer restoreStdinStdout(realin, realout)

	realdir, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get working dir. Args: %q, Field: %s, Error: %s", test.Args, test.Field, err)
		return
	}
	defer restoreWorkingDir(realdir)

	err = setupIOFieldTest(test)
	if err != nil {
		t.Errorf("Failed to setup test. Args: %q, Field: %s, Error: %s", test.Args, test.Field, err)
		return
	}

	cmd := New("test", spec)
	_, _, err = cmd.Decode(test.Args)
	if !test.Valid {
		if err == nil {
			t.Errorf("Expected error but none received. Args: %q", test.Args)
		}
		return
	}
	if err != nil {
		t.Errorf("Received unexpected decode error. Args: %q, Field: %s, Error: %s", test.Args, test.Field, err)
		return
	}

	err = validateIOFieldTest(spec, test)
	if err != nil {
		t.Errorf("Validation failed during IO field test. Args: %q, Field: %s, Error: %s", test.Args, test.Field, err)
		return
	}
}

func restoreStdinStdout(stdin *os.File, stdout *os.File) {
	os.Stdin = stdin
	os.Stdout = stdout
}

func restoreWorkingDir(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		panic(fmt.Sprintf("Failed to restore working dir. Dir: %q, Error: %s", dir, err))
	}
}

func setupIOFieldTest(test ioFieldTest) error {
	tmpdir, err := ioutil.TempDir("", "writ-iofieldtest")
	if err != nil {
		return err
	}
	err = os.Chdir(tmpdir)
	if err != nil {
		return err
	}

	for _, name := range test.InFiles {
		f, err := os.Create(name)
		if err != nil {
			return err
		}
		_, err = io.WriteString(f, ioTestText)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		if name == "stdin" {
			in, err := os.Open(name)
			if err != nil {
				return err
			}
			os.Stdin = in
		}
	}
	for _, name := range test.OutFiles {
		if name == "stdout" {
			out, err := os.Create(name)
			if err != nil {
				return err
			}
			os.Stdout = out
		}
	}
	return nil
}

func validateIOFieldTest(spec *ioFieldSpec, test ioFieldTest) error {
	err := spec.PerformIO()
	if err != nil {
		return err
	}
	for _, name := range test.OutFiles {
		in, err := os.Open(name)
		if err != nil {
			return err
		}
		bytes, err := ioutil.ReadAll(in)
		if err != nil {
			return err
		}
		if string(bytes) != ioTestText {
			return fmt.Errorf("Expected to read %q. Read %q instead.", ioTestText, string(bytes))
		}
		err = in.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
 * Test custom flag and option decoders
 */

type customTestFlag struct {
	val bool
}

func (d *customTestFlag) Decode(arg string) error {
	d.val = true
	return nil
}

type customTestOption struct {
	val string
}

func (d *customTestOption) Decode(arg string) error {
	if strings.HasPrefix(arg, "foo") {
		d.val = arg
		return nil
	} else {
		return fmt.Errorf("customTestOption values must begin with foo")
	}
}

type customDecoderFieldSpec struct {
	CustomFlag   customTestFlag   `flag:"f, flag" description:"a custom flag field"`
	CustomOption customTestOption `option:"o, opt" description:"a custom option field"`
}

var customDecoderFieldTests = []fieldTest{
	// Custom flag
	{Args: []string{"-f"}, Valid: true, Field: "CustomFlag", Value: customTestFlag{val: true}},
	{Args: []string{"--flag"}, Valid: true, Field: "CustomFlag", Value: customTestFlag{val: true}},
	{Args: []string{"--flag", "--flag"}, Valid: false}, // Plural must be set explicitly
	{Args: []string{"-ff"}, Valid: false},              // Plural must be set explicitly

	// Custom option
	{Args: []string{"-ofoobar"}, Valid: true, Field: "CustomOption", Value: customTestOption{val: "foobar"}},
	{Args: []string{"-o", "foobar"}, Valid: true, Field: "CustomOption", Value: customTestOption{val: "foobar"}},
	{Args: []string{"--opt", "foobar"}, Valid: true, Field: "CustomOption", Value: customTestOption{val: "foobar"}},
	{Args: []string{"--opt=foobar"}, Valid: true, Field: "CustomOption", Value: customTestOption{val: "foobar"}},
	{Args: []string{"-o", "puppies"}, Valid: false},
	{Args: []string{"-opuppies"}, Valid: false},
	{Args: []string{"-opt=puppies"}, Valid: false},
	{Args: []string{"-opt", "puppies"}, Valid: false},
	{Args: []string{"-o"}, Valid: false},
	{Args: []string{"--opt"}, Valid: false},
	{Args: []string{"--opt", "foobar", "-o", "foobar"}, Valid: false}, // Plural must be set explicitly
	{Args: []string{"-ofoobar", "-ofoobar"}, Valid: false},            // Plural must be set explicitly
}

func TestCustomDecoderFields(t *testing.T) {
	for _, test := range customDecoderFieldTests {
		spec := &customDecoderFieldSpec{}
		runFieldTest(t, spec, test)
	}
}

/*
 * Test basic field types
 */

type basicFieldSpec struct {
	Int     int     `option:"int" description:"An int option" placeholder:"INT"`
	Int8    int8    `option:"int8" description:"An int8 option" placeholder:"INT8"`
	Int16   int16   `option:"int16" description:"An int16 option" placeholder:"INT16"`
	Int32   int32   `option:"int32" description:"An int32 option" placeholder:"INT32"`
	Int64   int64   `option:"int64" description:"An int64 option" placeholder:"INT64"`
	Uint    uint    `option:"uint" description:"A uint option" placeholder:"UINT"`
	Uint8   uint8   `option:"uint8" description:"A uint8 option" placeholder:"UINT8"`
	Uint16  uint16  `option:"uint16" description:"A uint16 option" placeholder:"UINT16"`
	Uint32  uint32  `option:"uint32" description:"A uint32 option" placeholder:"UINT32"`
	Uint64  uint64  `option:"uint64" description:"A uint64 option" placeholder:"UINT64"`
	Float32 float32 `option:"float32" description:"A float32 option" placeholder:"FLOAT32"`
	Float64 float64 `option:"float64" description:"A float64 option" placeholder:"FLOAT64"`
	String  string  `option:"string" description:"A string option" placeholder:"STRING"`
}

var basicFieldTests = []fieldTest{
	// String
	{Args: []string{"--string", "1"}, Valid: true, Field: "String", Value: "1"},
	{Args: []string{"--string", "-1"}, Valid: true, Field: "String", Value: "-1"},
	{Args: []string{"--string", "+1"}, Valid: true, Field: "String", Value: "+1"},
	{Args: []string{"--string", "1.0"}, Valid: true, Field: "String", Value: "1.0"},
	{Args: []string{"--string", "0x01"}, Valid: true, Field: "String", Value: "0x01"},
	{Args: []string{"--string", "-"}, Valid: true, Field: "String", Value: "-"},
	{Args: []string{"--string", "-a"}, Valid: true, Field: "String", Value: "-a"},
	{Args: []string{"--string", "--"}, Valid: true, Field: "String", Value: "--"},
	{Args: []string{"--string", "--a"}, Valid: true, Field: "String", Value: "--a"},
	{Args: []string{"--string", ""}, Valid: true, Field: "String", Value: ""},
	{Args: []string{"--string", " "}, Valid: true, Field: "String", Value: " "},
	{Args: []string{"--string", " a"}, Valid: true, Field: "String", Value: " a"},
	{Args: []string{"--string", "a "}, Valid: true, Field: "String", Value: "a "},
	{Args: []string{"--string", "a b "}, Valid: true, Field: "String", Value: "a b "},
	{Args: []string{"--string", " a b"}, Valid: true, Field: "String", Value: " a b"},
	{Args: []string{"--string", "\n"}, Valid: true, Field: "String", Value: "\n"},
	{Args: []string{"--string", "\t"}, Valid: true, Field: "String", Value: "\t"},
	{Args: []string{"--string", "日本"}, Valid: true, Field: "String", Value: "日本"},
	{Args: []string{"--string", "-日本"}, Valid: true, Field: "String", Value: "-日本"},
	{Args: []string{"--string", "--日本"}, Valid: true, Field: "String", Value: "--日本"},
	{Args: []string{"--string", " 日本"}, Valid: true, Field: "String", Value: " 日本"},
	{Args: []string{"--string", "日本 "}, Valid: true, Field: "String", Value: "日本 "},
	{Args: []string{"--string", "日 本"}, Valid: true, Field: "String", Value: "日 本"},
	{Args: []string{"--string", "A relatively long string to make sure we aren't doing any silly truncation anywhere, since that would be bad..."}, Valid: true, Field: "String", Value: "A relatively long string to make sure we aren't doing any silly truncation anywhere, since that would be bad..."},
	{Args: []string{"--string", "a", "--string", "b"}, Valid: false},
	{Args: []string{"--string"}, Valid: false},

	// Int8
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: true, Field: "Int8", Value: int8(math.MinInt8)},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Int8", Value: int8(math.MaxInt8)},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int8", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int8", "1", "--int8", "2"}, Valid: false},
	{Args: []string{"--int8", "1.0"}, Valid: false},
	{Args: []string{"--int8", ""}, Valid: false},
	{Args: []string{"--int8"}, Valid: false},

	// Int16
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: true, Field: "Int16", Value: int16(math.MinInt8)},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: true, Field: "Int16", Value: int16(math.MinInt8 - 1)},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Int16", Value: int16(math.MaxInt8)},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Int16", Value: int16(math.MaxInt8 + 1)},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: true, Field: "Int16", Value: int16(math.MinInt16)},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Int16", Value: int16(math.MaxInt16)},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Int16", Value: int16(math.MaxUint8)},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Int16", Value: int16(math.MaxUint8 + 1)},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int16", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int16", "1", "--int16", "2"}, Valid: false},
	{Args: []string{"--int16", "1.0"}, Valid: false},
	{Args: []string{"--int16", ""}, Valid: false},
	{Args: []string{"--int16"}, Valid: false},

	// Int32
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: true, Field: "Int32", Value: int32(math.MinInt8)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: true, Field: "Int32", Value: int32(math.MinInt8 - 1)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Int32", Value: int32(math.MaxInt8)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Int32", Value: int32(math.MaxInt8 + 1)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: true, Field: "Int32", Value: int32(math.MinInt16)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: true, Field: "Int32", Value: int32(math.MinInt16 - 1)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Int32", Value: int32(math.MaxInt16)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Int32", Value: int32(math.MaxInt16 + 1)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: true, Field: "Int32", Value: int32(math.MinInt32)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: true, Field: "Int32", Value: int32(math.MaxInt32)},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Int32", Value: int32(math.MaxUint8)},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Int32", Value: int32(math.MaxUint8 + 1)},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Int32", Value: int32(math.MaxUint16)},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: true, Field: "Int32", Value: int32(math.MaxUint16 + 1)},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int32", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int32", "1", "--int32", "2"}, Valid: false},
	{Args: []string{"--int32", "1.0"}, Valid: false},
	{Args: []string{"--int32", ""}, Valid: false},
	{Args: []string{"--int32"}, Valid: false},

	// Int64
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: true, Field: "Int64", Value: int64(math.MinInt8)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: true, Field: "Int64", Value: int64(math.MinInt8 - 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt8)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt8 + 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: true, Field: "Int64", Value: int64(math.MinInt16)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: true, Field: "Int64", Value: int64(math.MinInt16 - 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt16)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt16 + 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: true, Field: "Int64", Value: int64(math.MinInt32)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: true, Field: "Int64", Value: int64(math.MinInt32 - 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt32)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt32 + 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: true, Field: "Int64", Value: int64(math.MinInt64)},
	{Args: []string{"--int64", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: true, Field: "Int64", Value: int64(math.MaxInt64)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Int64", Value: int64(math.MaxUint8)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Int64", Value: int64(math.MaxUint8 + 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Int64", Value: int64(math.MaxUint16)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: true, Field: "Int64", Value: int64(math.MaxUint16 + 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: true, Field: "Int64", Value: int64(math.MaxUint32)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: true, Field: "Int64", Value: int64(math.MaxUint32 + 1)},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int64", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int64", "1", "--int64", "2"}, Valid: false},
	{Args: []string{"--int64", "1.0"}, Valid: false},
	{Args: []string{"--int64", ""}, Valid: false},
	{Args: []string{"--int64"}, Valid: false},

	// Int
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: true, Field: "Int", Value: int(math.MinInt8)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: true, Field: "Int", Value: int(math.MinInt8 - 1)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Int", Value: int(math.MaxInt8)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Int", Value: int(math.MaxInt8 + 1)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: true, Field: "Int", Value: int(math.MinInt16)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: true, Field: "Int", Value: int(math.MinInt16 - 1)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Int", Value: int(math.MaxInt16)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Int", Value: int(math.MaxInt16 + 1)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: true, Field: "Int", Value: int(math.MinInt32)},
	{Args: []string{"--int", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: true, Field: "Int", Value: int(math.MaxInt32)},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Int", Value: int(math.MaxUint8)},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Int", Value: int(math.MaxUint8 + 1)},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Int", Value: int(math.MaxUint16)},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: true, Field: "Int", Value: int(math.MaxUint16 + 1)},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--int", "1", "--int", "2"}, Valid: false},
	{Args: []string{"--int", "1.0"}, Valid: false},
	{Args: []string{"--int", ""}, Valid: false},
	{Args: []string{"--int"}, Valid: false},

	// Uint8
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Uint8", Value: uint8(math.MaxInt8)},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Uint8", Value: uint8(math.MaxInt8 + 1)},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Uint8", Value: uint8(math.MaxUint8)},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--uint8", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--uint8", "1", "--uint8", "2"}, Valid: false},
	{Args: []string{"--uint8", "1.0"}, Valid: false},
	{Args: []string{"--uint8", ""}, Valid: false},
	{Args: []string{"--uint8"}, Valid: false},

	// Uint16
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxInt8)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxInt8 + 1)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxInt16)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxInt16 + 1)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxUint8)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxUint8 + 1)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Uint16", Value: uint16(math.MaxUint16)},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--uint16", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--uint16", "1", "--uint16", "2"}, Valid: false},
	{Args: []string{"--uint16", "1.0"}, Valid: false},
	{Args: []string{"--uint16", ""}, Valid: false},
	{Args: []string{"--uint16"}, Valid: false},

	// Uint32
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxInt8)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxInt8 + 1)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxInt16)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxInt16 + 1)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxInt32)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxInt32 + 1)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxUint8)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxUint8 + 1)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxUint16)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxUint16 + 1)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: true, Field: "Uint32", Value: uint32(math.MaxUint32)},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--uint32", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: false},
	{Args: []string{"--uint32", "1", "--uint32", "1"}, Valid: false},
	{Args: []string{"--uint32", "1.0"}, Valid: false},
	{Args: []string{"--uint32", ""}, Valid: false},
	{Args: []string{"--uint32"}, Valid: false},

	// Uint64
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt8)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt8 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt16)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt16 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt32)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt32 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MaxInt64))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt64)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxInt64+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxInt64 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint8)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint8 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint16)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint16 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint32)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint32+1))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint32 + 1)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", uint64(math.MaxUint64))}, Valid: true, Field: "Uint64", Value: uint64(math.MaxUint64)},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--uint64", fmt.Sprintf("%d", int64(math.MinInt64))}, Valid: false},
	{Args: []string{"--uint64", "1", "--uint64", "1"}, Valid: false},
	{Args: []string{"--uint64", "1.0"}, Valid: false},
	{Args: []string{"--uint64", ""}, Valid: false},
	{Args: []string{"--uint64"}, Valid: false},

	// Uint
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MaxInt8))}, Valid: true, Field: "Uint", Value: uint(math.MaxInt8)},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MaxInt8+1))}, Valid: true, Field: "Uint", Value: uint(math.MaxInt8 + 1)},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MaxInt16))}, Valid: true, Field: "Uint", Value: uint(math.MaxInt16)},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MaxInt16+1))}, Valid: true, Field: "Uint", Value: uint(math.MaxInt16 + 1)},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MaxInt32))}, Valid: true, Field: "Uint", Value: uint(math.MaxInt32)},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MaxInt32+1))}, Valid: true, Field: "Uint", Value: uint(math.MaxInt32 + 1)},
	{Args: []string{"--uint", fmt.Sprintf("%d", uint64(math.MaxUint8))}, Valid: true, Field: "Uint", Value: uint(math.MaxUint8)},
	{Args: []string{"--uint", fmt.Sprintf("%d", uint64(math.MaxUint8+1))}, Valid: true, Field: "Uint", Value: uint(math.MaxUint8 + 1)},
	{Args: []string{"--uint", fmt.Sprintf("%d", uint64(math.MaxUint16))}, Valid: true, Field: "Uint", Value: uint(math.MaxUint16)},
	{Args: []string{"--uint", fmt.Sprintf("%d", uint64(math.MaxUint16+1))}, Valid: true, Field: "Uint", Value: uint(math.MaxUint16 + 1)},
	{Args: []string{"--uint", fmt.Sprintf("%d", uint64(math.MaxUint32))}, Valid: true, Field: "Uint", Value: uint(math.MaxUint32)},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt8))}, Valid: false},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt8-1))}, Valid: false},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt16))}, Valid: false},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt16-1))}, Valid: false},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt32))}, Valid: false},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--uint", fmt.Sprintf("%d", int64(math.MinInt32-1))}, Valid: false},
	{Args: []string{"--uint", "1", "--uint", "2"}, Valid: false},
	{Args: []string{"--uint", "1.0"}, Valid: false},
	{Args: []string{"--uint", ""}, Valid: false},
	{Args: []string{"--uint"}, Valid: false},

	// Float32
	{Args: []string{"--float32", "-1.23"}, Valid: true, Field: "Float32", Value: float32(-1.23)},
	{Args: []string{"--float32", "4.56"}, Valid: true, Field: "Float32", Value: float32(4.56)},
	{Args: []string{"--float32", "-1.2e3"}, Valid: true, Field: "Float32", Value: float32(-1.2e3)},
	{Args: []string{"--float32", "4.5e6"}, Valid: true, Field: "Float32", Value: float32(4.5e6)},
	{Args: []string{"--float32", "-1.2E3"}, Valid: true, Field: "Float32", Value: float32(-1.2e3)},
	{Args: []string{"--float32", "4.5E6"}, Valid: true, Field: "Float32", Value: float32(4.5e6)},
	{Args: []string{"--float32", "-1.2e+3"}, Valid: true, Field: "Float32", Value: float32(-1.2e3)},
	{Args: []string{"--float32", "4.5e+6"}, Valid: true, Field: "Float32", Value: float32(4.5e6)},
	{Args: []string{"--float32", "-1.2E+3"}, Valid: true, Field: "Float32", Value: float32(-1.2e3)},
	{Args: []string{"--float32", "4.5E+6"}, Valid: true, Field: "Float32", Value: float32(4.5e6)},
	{Args: []string{"--float32", "-1.2e-3"}, Valid: true, Field: "Float32", Value: float32(-1.2e-3)},
	{Args: []string{"--float32", "4.5e-6"}, Valid: true, Field: "Float32", Value: float32(4.5e-6)},
	{Args: []string{"--float32", "-1.2E-3"}, Valid: true, Field: "Float32", Value: float32(-1.2e-3)},
	{Args: []string{"--float32", "4.5E-6"}, Valid: true, Field: "Float32", Value: float32(4.5e-6)},
	{Args: []string{"--float32", strconv.FormatFloat(math.SmallestNonzeroFloat32, 'f', -1, 64)}, Valid: true, Field: "Float32", Value: float32(math.SmallestNonzeroFloat32)},
	{Args: []string{"--float32", strconv.FormatFloat(math.MaxFloat32, 'f', -1, 64)}, Valid: true, Field: "Float32", Value: float32(math.MaxFloat32)},
	{Args: []string{"--float32", strconv.FormatFloat(math.MaxFloat32, 'f', -1, 64)}, Valid: true, Field: "Float32", Value: float32(math.MaxFloat32)},
	// XXX Skipped -- Not sure how to handle this!!
	{Args: []string{"--float32", strconv.FormatFloat(math.SmallestNonzeroFloat64, 'f', -1, 64)}, Field: "Float32", SkipReason: "Not sure how to handle the precision on this"},
	{Args: []string{"--float32", strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64)}, Valid: false},
	{Args: []string{"--float32", strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64)}, Valid: false},
	{Args: []string{"--float32", "1"}, Valid: true, Field: "Float32", Value: float32(1)},
	{Args: []string{"--float32", "-1"}, Valid: true, Field: "Float32", Value: float32(-1)},
	{Args: []string{"--float32", "1.0", "--float32", "2.0"}, Valid: false},
	{Args: []string{"--float32", ""}, Valid: false},
	{Args: []string{"--float32"}, Valid: false},

	// Float64
	{Args: []string{"--float64", "-1.23"}, Valid: true, Field: "Float64", Value: float64(-1.23)},
	{Args: []string{"--float64", "4.56"}, Valid: true, Field: "Float64", Value: float64(4.56)},
	{Args: []string{"--float64", "-1.2e3"}, Valid: true, Field: "Float64", Value: float64(-1.2e3)},
	{Args: []string{"--float64", "4.5e6"}, Valid: true, Field: "Float64", Value: float64(4.5e6)},
	{Args: []string{"--float64", "-1.2E3"}, Valid: true, Field: "Float64", Value: float64(-1.2e3)},
	{Args: []string{"--float64", "4.5E6"}, Valid: true, Field: "Float64", Value: float64(4.5e6)},
	{Args: []string{"--float64", "-1.2e+3"}, Valid: true, Field: "Float64", Value: float64(-1.2e3)},
	{Args: []string{"--float64", "4.5e+6"}, Valid: true, Field: "Float64", Value: float64(4.5e6)},
	{Args: []string{"--float64", "-1.2E+3"}, Valid: true, Field: "Float64", Value: float64(-1.2e3)},
	{Args: []string{"--float64", "4.5E+6"}, Valid: true, Field: "Float64", Value: float64(4.5e6)},
	{Args: []string{"--float64", "-1.2e-3"}, Valid: true, Field: "Float64", Value: float64(-1.2e-3)},
	{Args: []string{"--float64", "4.5e-6"}, Valid: true, Field: "Float64", Value: float64(4.5e-6)},
	{Args: []string{"--float64", "-1.2E-3"}, Valid: true, Field: "Float64", Value: float64(-1.2e-3)},
	{Args: []string{"--float64", "4.5E-6"}, Valid: true, Field: "Float64", Value: float64(4.5e-6)},
	{Args: []string{"--float64", strconv.FormatFloat(math.SmallestNonzeroFloat32, 'f', -1, 64)}, Valid: true, Field: "Float64", Value: float64(math.SmallestNonzeroFloat32)},
	{Args: []string{"--float64", strconv.FormatFloat(math.MaxFloat32, 'f', -1, 64)}, Valid: true, Field: "Float64", Value: float64(math.MaxFloat32)},
	{Args: []string{"--float64", strconv.FormatFloat(math.SmallestNonzeroFloat64, 'f', -1, 64)}, Valid: true, Field: "Float64", Value: float64(math.SmallestNonzeroFloat64)},
	{Args: []string{"--float64", strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64)}, Valid: true, Field: "Float64", Value: float64(math.MaxFloat64)},
	{Args: []string{"--float64", "1"}, Valid: true, Field: "Float64", Value: float64(1)},
	{Args: []string{"--float64", "-1"}, Valid: true, Field: "Float64", Value: float64(-1)},
	{Args: []string{"--float64", "1.0", "--float64", "2.0"}, Valid: false},
	{Args: []string{"--float64", ""}, Valid: false},
	{Args: []string{"--float64"}, Valid: false},
}

func TestBasicFields(t *testing.T) {
	for _, test := range basicFieldTests {
		spec := &basicFieldSpec{}
		runFieldTest(t, spec, test)
	}
}

/*
 * Test invalid specs
 */

var invalidSpecTests = []struct {
	Description string
	Spec        interface{}
}{
	// Invalid command specs
	{
		Description: "Commands must have a name 1",
		Spec: &struct {
			Command struct{} `command:","`
		}{},
	},
	{
		Description: "Commands must have a name 2",
		Spec: &struct {
			Command struct{} `command:" "`
		}{},
	},
	{
		Description: "Commands must have a single name",
		Spec: &struct {
			Command struct{} `command:"one,two"`
		}{},
	},
	{
		Description: "Command names cannot have a leading '-' prefix",
		Spec: &struct {
			Command struct{} `command:"-command"`
		}{},
	},
	{
		Description: "Command aliases cannot have a leading '-' prefix",
		Spec: &struct {
			Command struct{} `command:"command" alias:"-alias"`
		}{},
	},
	{
		Description: "Commands cannot have placeholders",
		Spec: &struct {
			Command struct{} `command:"command" placeholder:"PLACEHOLDER"`
		}{},
	},
	{
		Description: "Commands cannot have default values",
		Spec: &struct {
			Command struct{} `command:"command" default:"default"`
		}{},
	},
	{
		Description: "Commands cannot have env values",
		Spec: &struct {
			Command struct{} `command:"command" env:"ENV_VALUE"`
		}{},
	},
	{
		Description: "Command fields must be exported",
		Spec: &struct {
			command struct{} `command:"command"`
		}{},
	},
	{
		Description: "Command and alias names must be unique 1",
		Spec: &struct {
			Command struct{} `command:"foo" alias:"foo"`
		}{},
	},
	{
		Description: "Command and alias names must be unique 2",
		Spec: &struct {
			Command1 struct{} `command:"foo"`
			Command2 struct{} `command:"foo"`
		}{},
	},
	{
		Description: "Command and alias names must be unique 3",
		Spec: &struct {
			Command1 struct{} `command:"foo"`
			Command2 struct{} `command:"b" alias:"foo"`
		}{},
	},
	{
		Description: "Command and alias names must be unique 4",
		Spec: &struct {
			Command1 struct{} `command:"a" alias:"foo"`
			Command2 struct{} `command:"b" alias:"foo"`
		}{},
	},
	{
		Description: "Command specs must be a pointer to struct 1",
		Spec:        struct{}{},
	},
	{
		Description: "Command specs must be a pointer to struct 2",
		Spec:        42,
	},

	// Invalid option specs
	{
		Description: "Options cannot have aliases",
		Spec: &struct {
			Option int `option:"option" alias:"alias" description:"option with an alias"`
		}{},
	},
	{
		Description: "Options must have a name 1",
		Spec: &struct {
			Option int `option:"," description:"option with no name"`
		}{},
	},
	{
		Description: "Options must have a name 2",
		Spec: &struct {
			Option int `option:" " description:"option with no name"`
		}{},
	},
	{
		Description: "Long option names cannot have a leading '-' prefix",
		Spec: &struct {
			Option int `option:"-option" description:"leading dash prefix"`
		}{},
	},
	{
		Description: "Short option names cannot have a leading '-' prefix",
		Spec: &struct {
			Option int `option:"-o" description:"leading dash prefix"`
		}{},
	},
	{
		Description: "Option fields must be exported",
		Spec: &struct {
			option int `option:"option" description:"non-exported field"`
		}{},
	},
	{
		Description: "Bools cannot be options",
		Spec: &struct {
			Option bool `option:"b" description:"boolean option"`
		}{},
	},
	{
		Description: "Option names must be unique 1",
		Spec: &struct {
			Option1 int `option:"foo"`
			Option2 int `option:"foo"`
		}{},
	},
	{
		Description: "Option names must be unique 2",
		Spec: &struct {
			Option1 int `option:"a, foo"`
			Option2 int `option:"b, foo"`
		}{},
	},
	{
		Description: "Option names must be unique 3",
		Spec: &struct {
			Flag   bool `flag:"foo"`
			Option int  `option:"foo"`
		}{},
	},

	// Invalid flag specs
	{
		Description: "Flags cannot have aliases",
		Spec: &struct {
			Flag bool `flag:"flag" alias:"alias" description:"flag with an alias"`
		}{},
	},
	{
		Description: "Flags cannot have placeholders",
		Spec: &struct {
			Flag bool `flag:"flag" placeholder:"PLACEHOLDER" description:"placeholder on flag"`
		}{},
	},
	{
		Description: "Flags cannot have default values",
		Spec: &struct {
			Flag bool `flag:"flag" default:"default" description:"default on flag"`
		}{},
	},
	{
		Description: "Flags cannot have env values",
		Spec: &struct {
			Flag bool `flag:"flag" env:"ENV_VALUE" description:"env on flag"`
		}{},
	},
	{
		Description: "Flags must have a name 1",
		Spec: &struct {
			Flag bool `flag:"," description:"flag with no name"`
		}{},
	},
	{
		Description: "Flags must have a name 2",
		Spec: &struct {
			Flag bool `flag:" " description:"flag with no name"`
		}{},
	},
	{
		Description: "Long flag names cannot have a leading '-' prefix",
		Spec: &struct {
			Flag bool `flag:"-flag" description:"leading dash prefix"`
		}{},
	},
	{
		Description: "Short flag names cannot have a leading '-' prefix",
		Spec: &struct {
			Flag bool `flag:"-f" description:"leading dash prefix"`
		}{},
	},
	{
		Description: "Flag fields must be exported",
		Spec: &struct {
			flag int `flag:"flag" description:"non-exported field"`
		}{},
	},
	{
		Description: "Flag names must be unique 1",
		Spec: &struct {
			Flag1 bool `flag:"foo"`
			Flag2 bool `flag:"foo"`
		}{},
	},
	{
		Description: "Flag names must be unique 2",
		Spec: &struct {
			Flag1 bool `flag:"a, foo"`
			Flag2 bool `flag:"b, foo"`
		}{},
	},
	{
		Description: "Flag names must be unique 3",
		Spec: &struct {
			Flag   bool `flag:"foo"`
			Option int  `option:"foo"`
		}{},
	},

	// Invalid mixes of command, flag, and option
	{
		Description: "Commands cannot be options",
		Spec: &struct {
			Command struct{} `command:"command" option:"option" description:"command as option"`
		}{},
	},
	{
		Description: "Commands cannot be flags",
		Spec: &struct {
			Command struct{} `command:"command" flag:"flag" description:"command as flag"`
		}{},
	},
	{
		Description: "Options cannot be commands",
		Spec: &struct {
			Option int `option:"option" command:"command" description:"option as command"`
		}{},
	},
	{
		Description: "Options cannot be flags",
		Spec: &struct {
			Option int `option:"option" flag:"flag" description:"option as flag"`
		}{},
	},
	{
		Description: "Flags cannot be commands",
		Spec: &struct {
			Flag bool `flag:"flag" command:"command" description:"flag as command"`
		}{},
	},
	{
		Description: "Flags cannot be options",
		Spec: &struct {
			Flag bool `flag:"flag" option:"option" description:"flag as option"`
		}{},
	},
}

func TestInvalidSpecs(t *testing.T) {
	for _, test := range invalidSpecTests {
		err := newInvalidCommand(test.Spec)
		if err == nil {
			t.Errorf("Expected error creating spec, but none received.  Test: %s", test.Description)
			continue
		}
	}
}

func newInvalidCommand(spec interface{}) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			switch e := r.(type) {
			case commandError:
				err = e
			case optionError:
				err = e
			default:
				panic(e)
			}
		}
	}()
	New("test", spec)
	return nil
}

var invalidCommandTests = []struct {
	Description string
	Command     *Command
}{
	{
		Description: "Command names cannot begin with -",
		Command:     &Command{Name: "-command"},
	},
	{
		Description: "Command aliases cannot begin with -",
		Command:     &Command{Name: "command", Aliases: []string{"-alias"}},
	},
	{
		Description: "Command names cannot have spaces 1",
		Command:     &Command{Name: " command"},
	},
	{
		Description: "Command names cannot have spaces 2",
		Command:     &Command{Name: "command "},
	},
	{
		Description: "Command names cannot have spaces 3",
		Command:     &Command{Name: "command spaces"},
	},
	{
		Description: "Command aliases cannot begin with -",
		Command:     &Command{Name: "command", Aliases: []string{"-alias"}},
	},
	{
		Description: "Command aliases cannot have spaces 1",
		Command:     &Command{Name: "command", Aliases: []string{" alias"}},
	},
	{
		Description: "Command aliases cannot have spaces 2",
		Command:     &Command{Name: "command", Aliases: []string{"alias "}},
	},
	{
		Description: "Command aliases cannot have spaces 3",
		Command:     &Command{Name: "command", Aliases: []string{"alias spaces"}},
	},
}

func TestDirectCommandValidation(t *testing.T) {
	for _, test := range invalidCommandTests {
		err := checkInvalidCommand(test.Command)
		if err == nil {
			t.Errorf("Expected error validating command, but none received.  Test: %s", test.Description)
			continue
		}
	}
}

func checkInvalidCommand(cmd *Command) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			switch e := r.(type) {
			case commandError:
				err = e
			case optionError:
				err = e
			default:
				panic(e)
			}
		}
	}()
	cmd.validate()
	return nil
}

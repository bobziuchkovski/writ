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
	"testing"
)

/*
 * Much of the option testing occurs indirectly via command_test.go
 */

type noopDecoder struct{}

func (d noopDecoder) Decode(arg string) error { return nil }

var invalidOptionTests = []struct {
	Description string
	Option      *Option
}{
	{
		Description: "Option must have a name 1",
		Option:      &Option{Decoder: noopDecoder{}},
	},
	{
		Description: "Option must have a name 1",
		Option:      &Option{Names: []string{}, Decoder: noopDecoder{}},
	},
	{
		Description: "Option names cannot be blank",
		Option:      &Option{Names: []string{""}, Decoder: noopDecoder{}},
	},
	{
		Description: "Option names cannot begin with -",
		Option:      &Option{Names: []string{"-option"}, Decoder: noopDecoder{}},
	},
	{
		Description: "Option names cannot have spaces 1",
		Option:      &Option{Names: []string{" option"}, Decoder: noopDecoder{}},
	},
	{
		Description: "Option names cannot have spaces 2",
		Option:      &Option{Names: []string{"option "}, Decoder: noopDecoder{}},
	},
	{
		Description: "Option names cannot have spaces 3",
		Option:      &Option{Names: []string{"option spaces"}, Decoder: noopDecoder{}},
	},
	{
		Description: "Option must have a decoder",
		Option:      &Option{Names: []string{"option"}},
	},
}

func TestDirectOptionValidation(t *testing.T) {
	for _, test := range invalidOptionTests {
		err := checkInvalidOption(test.Option)
		if err == nil {
			t.Errorf("Expected error validating option, but none received.  Test: %s", test.Description)
			continue
		}
	}
}

func checkInvalidOption(opt *Option) (err error) {
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
	opt.validate()
	return nil
}

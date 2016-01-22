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
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	readerPtr      *io.Reader
	readCloserPtr  *io.ReadCloser
	writerPtr      *io.Writer
	writeCloserPtr *io.WriteCloser
	readerT        = reflect.TypeOf(readerPtr).Elem()
	readCloserT    = reflect.TypeOf(readCloserPtr).Elem()
	writerT        = reflect.TypeOf(writerPtr).Elem()
	writeCloserT   = reflect.TypeOf(writeCloserPtr).Elem()
)

type optionError struct {
	err error
}

func (e optionError) Error() string {
	return e.err.Error()
}

// panicOption reports invalid use of the Option type
func panicOption(format string, values ...interface{}) {
	e := optionError{fmt.Errorf(format, values...)}
	panic(e)
}

// Option is used to specify program options and flags.  Options are
// created implicitly when New() parses a command spec.  They may also be
// created directly if added flexibility or runtime dynamism is required.
type Option struct {
	// Required
	Names   []string      // All of the names for the Option, long and short alike
	Decoder OptionDecoder // The Decoder is responsible for parsing the Option's arguments

	// Optional
	Flag   bool // If Flag is set, the Option takes no arguments
	Plural bool // If Plural is set, the Option may be specified multiple times

	// Optional; used for help output only
	Description string
	Placeholder string
}

// ShortNames returns a filtered slice of the names that are exactly one rune in length
func (o *Option) ShortNames() []string {
	var short []string
	for _, n := range o.Names {
		if len([]rune(n)) == 1 {
			short = append(short, n)
		}
	}
	return short
}

// LongNames returns a filtered slice of the names that are longer than one rune in length
func (o *Option) LongNames() []string {
	var long []string
	for _, n := range o.Names {
		if len([]rune(n)) > 1 {
			long = append(long, n)
		}
	}
	return long
}

func (o *Option) String() string {
	var short, long []string
	for _, s := range o.ShortNames() {
		short = append(short, "-"+s)
	}
	for _, l := range o.LongNames() {
		long = append(long, "--"+l)
	}
	return strings.Join(append(short, long...), "/")
}

func (o *Option) validate() {
	if len(o.Names) == 0 {
		panicOption("Options require at least one name: %#v", o)
	}
	for _, n := range o.Names {
		if strings.HasPrefix(n, "-") {
			panicOption("Option names cannot begin with '-' (option %s)", n)
		}
		if n == "" {
			panicOption("Option names cannot be blank: %#v", o)
		}
	}
	if o.Decoder == nil {
		panicOption("Option decoder cannot be nil (option %s)", o.String())
	}
}

// OptionDecoder is the interface used for decoding Option arguments.
// Every Option must have an OptionDecoder assigned.  New() constructs
// and assigns OptionDecoders automatically for supported field types.
type OptionDecoder interface {
	Decode(arg string) error
}

type decoderFunc func(rval reflect.Value, arg string) error

func decodeInt(rval reflect.Value, arg string) error {
	v, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return err
	}
	if rval.OverflowInt(v) {
		return fmt.Errorf("value %d would overflow %s", v, rval.Kind())
	}
	rval.Set(reflect.ValueOf(v).Convert(rval.Type()))
	return nil
}

func decodeUint(rval reflect.Value, arg string) error {
	v, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return err
	}
	if rval.OverflowUint(v) {
		return fmt.Errorf("value %d would overflow %s", v, rval.Kind())
	}
	rval.Set(reflect.ValueOf(v).Convert(rval.Type()))
	return nil
}

func decodeFloat(rval reflect.Value, arg string) error {
	v, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	if rval.OverflowFloat(v) {
		return fmt.Errorf("value %f would overflow %s", v, rval.Kind())
	}
	rval.Set(reflect.ValueOf(v).Convert(rval.Type()))
	return nil
}

func decodeString(rval reflect.Value, arg string) error {
	rval.Set(reflect.ValueOf(arg))
	return nil
}

func getDecoderFunc(kind reflect.Kind) decoderFunc {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return decodeInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return decodeUint
	case reflect.Float32, reflect.Float64:
		return decodeFloat
	case reflect.String:
		return decodeString
	default:
		return nil
	}
}

// NewOptionDecoder builds an OptionDecoder for supported value types.  The val
// parameter must be a pointer to one of the supported types, or else
// NewOptionDecoder will panic.
//
// Supported types
//
// 		int, int8, int16, int32, int64, uint, uint8, iunt16, uint32, uint64
//		float32, float64
//		string, []string
//		map[string]string
//			Argument must be in key=value format.
//		io.Reader, io.ReadCloser
//			Argument must be a path to an existing file, or "-" to specify os.Stdin
//		io.Writer, io.WriteCloser
//			Argument will be used to create a new file, or "-" to specify os.Stdout.
//			If a file already exists at the path specified, it will be overwritten.
func NewOptionDecoder(val interface{}) OptionDecoder {
	// TODO: Check for nil val?
	// TODO: Cleanup
	rval := reflect.ValueOf(val)
	if rval.Kind() != reflect.Ptr {
		panicOption("NewDecoder must be called on a pointer")
	}
	elem := rval.Elem()

	// Not sure if this check is necessary or not....
	if !elem.CanSet() {
		panicOption("NewDecoder value must be settable")
	}
	etype := elem.Type()
	ekind := elem.Kind()

	var decoder OptionDecoder
	if etype == readerT || etype == readCloserT {
		decoder = inputDecoder{elem}
	} else if etype == writerT || etype == writeCloserT {
		decoder = outputDecoder{elem}
	} else if ekind == reflect.Slice && etype.Elem().Kind() == reflect.String {
		decoder = stringSliceDecoder{rval.Interface().(*[]string)}
	} else if ekind == reflect.Map && etype.Key().Kind() == reflect.String && etype.Elem().Kind() == reflect.String {
		decoder = stringMapDecoder{rval.Interface().(*map[string]string)}
	} else {
		decoderFunc := getDecoderFunc(ekind)
		if decoderFunc != nil {
			decoder = basicDecoder{elem, decoderFunc}
		}
	}
	if decoder == nil {
		panicOption("no option decoder available for type %s", rval.Type())
	}
	return decoder
}

type basicDecoder struct {
	rval        reflect.Value
	decoderFunc decoderFunc
}

func (d basicDecoder) Decode(arg string) error {
	return d.decoderFunc(d.rval, arg)
}

type stringSliceDecoder struct {
	value *[]string
}

func (d stringSliceDecoder) Decode(arg string) error {
	*d.value = append(*d.value, arg)
	return nil
}

type stringMapDecoder struct {
	value *map[string]string
}

func (d stringMapDecoder) Decode(arg string) error {
	keyval := strings.SplitN(arg, "=", 2)
	if len(keyval) != 2 {
		return fmt.Errorf("argument %q is not in key=value format", arg)
	}
	if *d.value == nil {
		*d.value = make(map[string]string)
	}
	(*d.value)[keyval[0]] = keyval[1]
	return nil
}

type inputDecoder struct {
	rval reflect.Value
}

func (d inputDecoder) Decode(arg string) error {
	var err error
	var f *os.File
	if arg == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(arg)
	}
	if err != nil {
		return err
	}
	d.rval.Set(reflect.ValueOf(f).Convert(d.rval.Type()))
	return nil
}

type outputDecoder struct {
	rval reflect.Value
}

func (d outputDecoder) Decode(arg string) error {
	var err error
	var f *os.File
	if arg == "-" {
		f = os.Stdout
	} else {
		f, err = os.Create(arg)
	}
	if err != nil {
		return err
	}
	d.rval.Set(reflect.ValueOf(f).Convert(d.rval.Type()))
	return nil
}

func (d flagAccumulator) Decode(arg string) error {
	*d.value++
	return nil
}

// NewFlagDecoder builds an OptionDecoder for boolean flag values.  The boolean
// value is set when the option is decoded.
func NewFlagDecoder(val *bool) OptionDecoder {
	if val == nil {
		panicOption("NewFlagDecoder called with a nil pointer")
	}
	return flagDecoder{val}
}

type flagDecoder struct {
	value *bool
}

func (d flagDecoder) Decode(arg string) error {
	*d.value = true
	return nil
}

// NewFlagAccumulator builds an OptionDecoder for int flag values.  The int value
// is incremented every time the option is decoded.
func NewFlagAccumulator(val *int) OptionDecoder {
	return flagAccumulator{val}
}

type flagAccumulator struct {
	value *int
}

// OptionDefaulter is used to initialize default option values.  If an
// OptionDecoder implemented the OptionDefaulter interface, its SetDefault()
// method is called prior to decoding options.
type OptionDefaulter interface {
	SetDefault()
}

// NewDefaulter builds an OptionDecoder that implements the OptionDefaulter
// interface.  The value for defaultArg is supplied to the underlying
// OptionDecoder's Decode() method when SetDefault() is called.
// If the value fails to decode, the OptionDefaulter will panic.
func NewDefaulter(d OptionDecoder, defaultArg string) OptionDecoder {
	return defaulter{d, defaultArg}
}

type defaulter struct {
	OptionDecoder
	defaultArg string
}

func (d defaulter) SetDefault() {
	err := d.Decode(d.defaultArg)
	if err != nil {
		// Default values should be known correct values, so we panic on error
		panicOption("error setting default value: decoder rejected arg %q", d.defaultArg)
	}
}

// NewEnvDefaulter builds an OptionDecoder that implements the OptionDefaulter
// interface.  The key's value is used to specify an environment variable.
// If the environment variable is set, the value is supplied to the underlying
// OptionDecoder's Decode() method when SetDefault() is called.s
// If the environment variable isn't set or fails to decode, the underlying
// OptionDecoder's SetDefault() method will be invoked (if the Decoder
// implements the OptionDefaulter interface).  Otherwise, no action is taken.
func NewEnvDefaulter(d OptionDecoder, key string) OptionDecoder {
	return envDefaulter{d, key}
}

type envDefaulter struct {
	OptionDecoder
	key string
}

func (d envDefaulter) SetDefault() {
	val := os.Getenv(d.key)
	if val != "" {
		err := d.Decode(val)
		if err == nil {
			return
		}
	}

	defaulter, ok := d.OptionDecoder.(OptionDefaulter)
	if ok {
		defaulter.SetDefault()
	}
}

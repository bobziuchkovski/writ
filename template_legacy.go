// +build !go1.6

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
	"text/template"
)

var defaultTemplate = template.Must(template.New("Help").Funcs(templateFuncs).Parse(HelpText))

// HelpText is used by Command.WriteHelp() and Command.ExitHelp() to generate
// help content.
//
// This copy of the template is used when compiling with go < 1.6.
// It's a bit ugly due to the lack of whitespace control with Go < 1.6.
// See template.go for the go1.6+ help template.
const HelpText = `{{/*
*/}}{{template "Main" .}}{{/*

*/}}{{define "Main"}}{{/*
*/}}{{template "Usage" .}}{{/*
*/}}{{template "Header" .}}{{/*
*/}}{{template "Body" .}}{{/*
*/}}{{template "Footer" .}}{{/*
*/}}{{end}}{{/*

*/}}{{define "Usage"}}{{/*
*/}}{{with .Help.Usage}}{{.}}{{"\n"}}{{end}}{{/*
*/}}{{end}}{{/*

*/}}{{define "Header"}}{{with .Help.Header}}{{.}}{{"\n"}}{{end}}{{end}}{{/*

*/}}{{define "Body"}}{{/*
*/}}{{template "OptionGroups" .}}{{/*
*/}}{{template "CommandGroups" .}}{{/*
*/}}{{end}}{{/*

*/}}{{define "OptionGroups"}}{{/*
*/}}{{with .Help.OptionGroups}}{{/*
*/}}{{range .}}{{template "OptionGroup" .}}{{end}}{{/*
*/}}{{end}}{{/*
*/}}{{end}}{{/*

*/}}{{define "OptionGroup"}}{{/*
*/}}{{"\n"}}{{/*
*/}}{{with .Header}}{{.}}{{"\n"}}{{end}}{{/*
*/}}{{with .Options}}{{/*
*/}}{{range .}}{{template "OptionHelp" .}}{{end}}{{/*
*/}}{{end}}{{/*
*/}}{{with .Footer}}{{.}}{{"\n"}}{{end}}{{/*
*/}}{{end}}{{/*

*/}}{{define "OptionHelp"}}{{formatOption .}}{{"\n"}}{{end}}{{/*

*/}}{{define "CommandGroups"}}{{/*
*/}}{{with .Help.CommandGroups}}{{/*
*/}}{{range .}}{{template "CommandGroup" .}}{{end}}{{/*
*/}}{{end}}{{/*
*/}}{{end}}{{/*

*/}}{{define "CommandGroup"}}{{/*
*/}}{{"\n"}}{{/*
*/}}{{with .Header}}{{.}}{{"\n"}}{{end}}{{/*
*/}}{{with .Commands}}{{/*
*/}}{{range .}}{{template "CommandHelp" .}}{{end}}{{/*
*/}}{{end}}{{/*
*/}}{{with .Footer}}{{.}}{{"\n"}}{{end}}{{/*
*/}}{{end}}{{/*

*/}}{{define "CommandHelp"}}{{formatCommand .}}{{"\n"}}{{end}}{{/*

*/}}{{define "Footer"}}{{with .Help.Footer}}{{"\n"}}{{.}}{{"\n"}}{{end}}{{end}}`

// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package writ_test

import (
	"github.com/ziuchkovski/writ"
	"os"
	"runtime"
)

type Config struct {
	help      bool
	verbosity int

	// A hidden flag
	tolerance float32

	// A dynamically added option for Mac OS
	useQuartz bool
}

// This example demonstrates explicit Command and Option construction
// without the use of writ.New()
func Example_explicit() {
	config := &Config{}
	cmd := &writ.Command{Name: "explicit"}
	cmd.Options = []*writ.Option{
		&writ.Option{
			Names:       []string{"h", "help"},
			Description: "Display this help text and exit",
			Decoder:     writ.NewFlagDecoder(&config.help),
			Flag:        true,
		},
		&writ.Option{
			Names:       []string{"v"},
			Description: "Increase verbosity; may be specified more than once",
			Decoder:     writ.NewFlagAccumulator(&config.verbosity),
			Flag:        true,
			Plural:      true,
		},
		&writ.Option{
			Names:       []string{"t", "tolerance"},
			Description: "Set the tolerance level (from 0.0 - 1.0)",
			Placeholder: "DECIMAL",
			Decoder:     writ.NewOptionDecoder(&config.tolerance),
		},
	}

	cmd.Help = writ.Help{
		Usage:  "Usage: explicit [OPTION]...",
		Header: "Explicit demonstrates explicit commands and options",
		Footer: "This method is flexible but more verbose than using writ.New()",
	}
	general := cmd.GroupOptions("help", "v")
	general.Header = "General Options:"
	cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, general)

	if runtime.GOOS == "darwin" {
		cmd.Options = append(cmd.Options, &writ.Option{
			Names:       []string{"use-quartz"},
			Description: "Use Quartz display on Mac",
			Decoder:     writ.NewFlagDecoder(&config.useQuartz),
			Flag:        true,
		})
		platform := cmd.GroupOptions("use-quartz")
		platform.Header = "Platform Options:"
		cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, platform)
	}

	// Decode the options
	_, _, err := cmd.Decode(os.Args[1:])
	if err != nil || config.help {
		cmd.ExitHelp(err)
	}
}

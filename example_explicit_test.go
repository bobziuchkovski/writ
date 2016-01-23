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
	useQuartz bool
}

// This example demonstrates explicit Command and Option creation,
// along with explicit option grouping.  It checks the host platform
// and dynamically adds a --with-quartz flag if the example is run on
// Mac OS.  The same result could be achieved by using writ.New() to
// construct a Command, and then adding the platform-specific option
// to the resulting Command directly.
func Example_explicit() {
	config := &Config{}
	cmd := &writ.Command{Name: "explicit"}
	cmd.Options = []*writ.Option{
		{
			Names:       []string{"h", "help"},
			Description: "Display this help text and exit",
			Decoder:     writ.NewFlagDecoder(&config.help),
			Flag:        true,
		},
		{
			Names:       []string{"v"},
			Description: "Increase verbosity; may be specified more than once",
			Decoder:     writ.NewFlagAccumulator(&config.verbosity),
			Flag:        true,
			Plural:      true,
		},
	}

	general := cmd.GroupOptions("help", "v")
	general.Header = "General Options:"
	cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, general)

	// Dynamically add --with-quartz on Mac OS
	if runtime.GOOS == "darwin" {
		cmd.Options = append(cmd.Options, &writ.Option{
			Names:       []string{"with-quartz"},
			Description: "Use Quartz display on Mac",
			Decoder:     writ.NewFlagDecoder(&config.useQuartz),
			Flag:        true,
		})
		platform := cmd.GroupOptions("with-quartz")
		platform.Header = "Platform Options:"
		cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, platform)
	}

	// Decode the options
	_, _, err := cmd.Decode(os.Args[1:])
	if err != nil || config.help {
		cmd.ExitHelp(err)
	}
}

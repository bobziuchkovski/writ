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
	help       bool
	verbosity  int
	bootloader string
}

// This example demonstrates explicit Command and Option creation,
// along with explicit option grouping.  It checks the host platform
// and dynamically adds a --bootloader option if the example is run on
// Linux.  The same result could be achieved by using writ.New() to
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

	// Dynamically add --bootloader on Linux
	if runtime.GOOS == "linux" {
		cmd.Options = append(cmd.Options, &writ.Option{
			Names:       []string{"bootloader"},
			Description: "Use the specified bootloader (grub, grub2, or lilo)",
			Decoder:     writ.NewOptionDecoder(&config.bootloader),
			Placeholder: "NAME",
		})
		platform := cmd.GroupOptions("bootloader")
		platform.Header = "Platform Options:"
		cmd.Help.OptionGroups = append(cmd.Help.OptionGroups, platform)
	}

	// Decode the options
	_, _, err := cmd.Decode(os.Args[1:])
	if err != nil || config.help {
		cmd.ExitHelp(err)
	}
}

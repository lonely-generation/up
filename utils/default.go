// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package utils

import "github.com/fatih/color"

var (
	defaults map[string]string = map[string]string{
		"TaskDir":    ".",
		"TaskFile":   "task",
		"FlowDir":    ".",
		"FlowFile":   "flow",
		"Verbose":    "v",
		"ConfigDir":  ".",
		"ConfigFile": "config",
	}
	vvvvv_color_printf  = color.Magenta
	verror_color_printf = color.Red
	msg_color_printf    = color.Yellow
	himsg_color_printf  = color.HiWhite
	msg_color_sprintf   = color.YellowString
	dryrun_color_print  = color.Cyan
)


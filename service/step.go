// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package service

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/stephencheng/up/interface/impl/funcs"
	u "github.com/stephencheng/up/utils"
)

type Step struct {
	Do   interface{} //FuncImpl
	Func string
	Desc string
}

func (step *Step) Exec() {
	cmdCnt := len(step.Do.([]interface{}))
	//u.P("cmd count:", cmdCnt)
	if cmdCnt > 1 {
		var cmds funcs.ShellCmds
		err := ms.Decode(step.Do, &cmds)
		u.LogError("e:", err)
		for idx, cmd := range cmds {
			u.Pfv("    cmd(%2d): %+v\n", idx+1, cmd)
			u.P("      exec result:", funcs.RunCmd(cmd))
		}
	} else {
		var cmd string
		err := ms.Decode(step.Do, &cmd)
		u.LogError("err:", err)
		u.P("      exec result:", funcs.RunCmd(cmd))
	}

}

type Steps []Step

func (steps *Steps) Exec() {

	for idx, step := range *steps {
		u.Pfvvvv("  step(%3d): %s\n", idx+1, u.PP(step))
		//u.Pfvvvv("%+v | length: %d\n", step.Do, len(step.Do.([]interface{})))
		step.Exec()
	}

}


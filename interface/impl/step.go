// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package impl

import (
	"github.com/imdario/mergo"
	"github.com/mohae/deepcopy"
	ic "github.com/stephencheng/up/interface"
	"github.com/stephencheng/up/model/cache"
	u "github.com/stephencheng/up/utils"
	ee "github.com/stephencheng/up/utils/error"
	//"gopkg.in/yaml.v2"
	"reflect"
	"strconv"
)

type Step struct {
	Name  string
	Do    interface{} //FuncImpl
	Func  string
	Vars  cache.Cache
	Dvars cache.Dvars
	Desc  string
	Reg   string
	Flags []string //ignore_error |
	If    string
	//Loop  *[]interface{}
	Loop interface{}
}

//this is final merged exec vars the individual step will use
//this step will merge the vars with the caller's stack vars
func (step *Step) GetExecVarsWithRefOverrided(funcname string, loopItem *LoopItem) *cache.Cache {
	vars := step.getRuntimeExecVars(funcname)
	callerVars := TaskStack.GetTop().(*TaskRuntimeContext).CallerVars
	//u.Ptmpdebug("callerVars", callerVars)

	if callerVars != nil {
		mergo.Merge(vars, callerVars, mergo.WithOverride)
	}

	if loopItem != nil {
		vars.Put("loopitem", loopItem.Item)
		vars.Put("loopindex", loopItem.Index)
		vars.Put("loopindex1", loopItem.Index1)
	}
	u.Ppmsgvvvvhint("overall final exec vars:", vars)
	return vars
}

/*
merge localvars to above RuntimeVarsAndDvarsMerged to get final runtime exec vars
the localvars is the vars in the step
*/
func (step *Step) getRuntimeExecVars(mark string) *cache.Cache {
	var execvars cache.Cache

	execvars = deepcopy.Copy(*cache.RuntimeVarsAndDvarsMerged).(cache.Cache)

	if step.Vars != nil {
		mergo.Merge(&execvars, step.Vars, mergo.WithOverride)

		u.Pfvvvv("current exec runtime[%s] vars:", mark)
		u.Ppmsgvvvv(execvars)
		u.Dvvvvv(execvars)
	}

	localVarsMergedWithDvars := cache.VarsMergedWithDvars("local", &step.Vars, &step.Dvars, &execvars)

	if localVarsMergedWithDvars.Len() > 0 {
		mergo.Merge(&execvars, localVarsMergedWithDvars, mergo.WithOverride)
	}

	return &execvars
}

type LoopItem struct {
	Index  int
	Index1 int
	Item   interface{}
}

func (step *Step) Exec() {
	var action ic.Do
	//u.Ptmpdebug("step debug", step)

	var bizErr *ee.Error = ee.New()
	var stepExecVars *cache.Cache

	plainExecVars := step.GetExecVarsWithRefOverrided("get plain exec vars", nil)

	routeFuncType := func(loopItem *LoopItem) {
		stepExecVars = step.GetExecVarsWithRefOverrided("step exec", loopItem)

		switch step.Func {
		case FUNC_SHELL:
			funcAction := ShellFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = ic.Do(&funcAction)

		case FUNC_TASK_REF:
			funcAction := TaskRefFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = ic.Do(&funcAction)

		case FUNC_NOOP:
			funcAction := NoopFuncAction{
				Do:   step.Do,
				Vars: stepExecVars,
			}
			action = ic.Do(&funcAction)

		default:
			u.InvalidAndExit("Step dispatch", "func name is not recognised and implemented")
			bizErr.Mark = "func name not implemented"
		}
	}

	dryRunOrContinue := func() {
		//example to stop further steps
		//f := u.MustConditionToContinueFunc(func() bool {
		//	return action != nil
		//})
		//
		//u.DryRunOrExit("Step Exec", f, "func name must be valid")

		alloweErrors := []string{
			"func name not implemented",
		}

		DryRunAndSkip(
			bizErr.Mark,
			alloweErrors,
			ContinueFunc(
				func() {
					if step.Loop != nil {
						func() {
							if reflect.TypeOf(step.Loop).Kind() == reflect.String {

								//toJsonTmp:="{{toJson .}}"
								//cache.Render("{{toJson .}}", plainExecVars)

								loopVarName := cache.Render(step.Loop.(string), plainExecVars)

								loopObj := plainExecVars.Get(loopVarName)
								if reflect.TypeOf(loopObj).Kind() == reflect.Slice {
									for idx, item := range loopObj.([]interface{}) {
										routeFuncType(&LoopItem{idx, idx + 1, item})
										action.Adapt()
										action.Exec()
									}

								} else {
									u.InvalidAndExit("evaluate loop var", "loop var is not a array/list/slice")
								}
							} else if reflect.TypeOf(step.Loop).Kind() == reflect.Slice {
								for idx, item := range step.Loop.([]interface{}) {
									routeFuncType(&LoopItem{idx, idx + 1, item})
									action.Adapt()
									action.Exec()
								}

							} else {
								u.InvalidAndExit("evaluate loop items", "please either use a list or a template evaluation which could result in a value of a list")
							}
						}()

					} else {
						routeFuncType(nil)
						action.Adapt()
						action.Exec()

					}

				}),
			nil,
		)
	}

	func() {
		if step.If != "" {
			IfEval := cache.Render(step.If, stepExecVars)
			goahead, err := strconv.ParseBool(IfEval)
			u.LogErrorAndExit("evaluate condition", err, "please fix if condition evaluation")
			if goahead {
				dryRunOrContinue()
			} else {
				u.Pvvvv("condition failed, skip executing step", step.Name)
			}
		} else {
			dryRunOrContinue()
		}

	}()

}

type Steps []Step

func (steps *Steps) Exec() {

	for idx, step := range *steps {
		u.Pf("step(%3d):\n", idx+1)
		//u.Pfvvvv("  step(%3d): %s\n", idx+1, u.Sppmsg(step))
		u.Ppmsgvvvv(step)

		execStep := func() {
			rtContext := StepRuntimeContext{
				Stepname: step.Name,
				Flags:    &step.Flags,
			}
			StepStack.Push(&rtContext)

			step.Exec()

			result := StepStack.GetTop().(*StepRuntimeContext).Result
			taskname := TaskStack.GetTop().(*TaskRuntimeContext).Taskname
			if u.Contains([]string{FUNC_SHELL, FUNC_TASK_REF}, step.Func) {
				if step.Reg == "auto" {
					cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Name), result.Output)
				} else if step.Reg != "" {
					cache.RuntimeVarsAndDvarsMerged.Put(u.Spf("register_%s_%s", taskname, step.Reg), result.Output)
				} else {
					if step.Func == FUNC_SHELL {
						cache.RuntimeVarsAndDvarsMerged.Put("last_task_result", result)
					}
				}
			}
			StepStack.Pop()
		}

		execStep()

	}

}


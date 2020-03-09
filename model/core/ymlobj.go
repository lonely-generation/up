// Ultimate Provisioner: UP cmd
// Copyright (c) 2019 Stephen Cheng and contributors

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package core

import (
	u "github.com/stephencheng/up/utils"
	yq "github.com/stephencheng/yq/v3/cmd"
	"gopkg.in/yaml.v2"
	"strings"
)

func ObjToYaml(obj interface{}) string {
	ymlbytes, err := yaml.Marshal(&obj)
	u.LogErrorAndExit("obj to yaml converstion", err, "yml convesion failed")
	return string(ymlbytes)
}

func YamlToObj(srcyml string) interface{} {
	obj := new(interface{})
	err := yaml.Unmarshal([]byte(srcyml), obj)
	u.LogErrorAndContinue("yml to object:", err, "please validate the ymal content")
	return obj
}

/*
obj is a cache item
path format: a.b.c(name=fr*).value
prefix will be used to get the obj, rest will be used as yq path
*/
func GetSubObject(cache *Cache, path string, collect bool) interface{} {
	//obj -> yml -> yq to get node in yml -> obj
	elist := strings.Split(path, ".")
	func() {
		if elist[0] == "" {
			u.InvalidAndExit("yml path validation", "path format is not correct, use format: a.b.c(name=fr*).value")
		}
	}()
	yqpath := strings.Join(elist[1:], ".")

	cacheKey := elist[0]
	obj := cache.Get(cacheKey)
	ymlstr := ObjToYaml(obj)
	u.Dvvvvv("sub yml str")
	u.Dvvvvv(ymlstr)
	yqresult, err := yq.UpReadYmlStr(ymlstr, yqpath, u.CoreConfig.Verbose, collect)
	u.LogErrorAndContinue("parse sub element in yml", err, u.Spf("please ensure yml query path: %s", yqpath))
	obj = YamlToObj(yqresult)
	return obj
}


// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package version

import "strings"

const Maj = "0"
const Min = "1"
const Fix = "0"

const Version = Maj + "." + Min + "." + Fix

var commitVer string

func GetVersion() string {
	return Version
}

func GetCommitVersion() string {
	if len(commitVer) < 6 {
		return Version + "-"
	}
	return Version + "-" + commitVer[:6]
}

/*=======================unholy separator===========================*/

var (
	app_name string
)

func InitNodeInfo(app string) {
	if len(app_name) > 0 {
		return
	}
	if slc := strings.Split(app, "-"); len(slc) > 1 {
		app_name = slc[1]
	} else {
		app_name = app
	}
}

func AppName() string {
	return app_name
}

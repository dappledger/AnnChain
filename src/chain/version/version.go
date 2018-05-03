/*
 * This file is part of The AnnChain.
 *
 * The AnnChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The AnnChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The www.annchain.io.  If not, see <http://www.gnu.org/licenses/>.
 */


package version

import "strings"

const Maj = "0"
const Min = "6"
const Fix = "0"

const Version = Maj + "." + Min + "." + Fix

var commitVer string

func GetVersion() string {
	return Version
}

func GetCommitVersion() string {
	return Version + "-" + commitVer
}

/*=======================  unholy separator  ===========================*/

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

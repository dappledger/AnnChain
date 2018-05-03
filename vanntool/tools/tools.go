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


package tools

import (
	"bytes"
	"strings"

	"github.com/BurntSushi/toml"
)

func EncodeToToml(inputs interface{}) (string, error) {
	var firstBuffer bytes.Buffer
	e := toml.NewEncoder(&firstBuffer)
	err := e.Encode(inputs)
	if err != nil {
		return "", err
	}
	return firstBuffer.String(), nil
}

func SplitTo2(str, split string) (str1 string, str2 string) {
	strSlc := strings.Split(str, split)
	if len(strSlc) > 0 {
		str1 = strSlc[0]
		if len(strSlc) > 1 {
			str2 = strSlc[1]
		}
	}
	return
}

func ParseRet(str string) string {
	return strings.Replace(str, "\\n", "<r/>", -1)
}

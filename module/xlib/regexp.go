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


package xlib

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexp_NotNumLetterUnderline *regexp.Regexp
)

func init() {
	//regexp_NotNumLetterUnderline = regexp.MustCompile(`^[0-9a-zA-Z_]+`)
	regexp_NotNumLetterUnderline = regexp.MustCompile(`[\W]+`)
}

func OnlyNumLetterUnderline(str string) bool {
	return regexp_NotNumLetterUnderline.FindStringIndex(str) == nil
}

func CheckNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

func CheckIPAddrSlc(str string) bool {
	addrSlc := strings.Split(str, ",")
	if len(addrSlc) == 0 {
		return false
	}
	for i := range addrSlc {
		if _, err := net.ResolveTCPAddr("tcp", addrSlc[i]); err != nil {
			return false
		}
	}
	return true
}

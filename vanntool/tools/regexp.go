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
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/dappledger/AnnChain/vanntool/def"
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

func CheckIPAddr(prefix, addr string) error {
	switch prefix {
	case "tcp":
	default:
		return fmt.Errorf("prefix only support tcp")
	}
	_, err := net.ResolveTCPAddr(prefix, addr)
	return err
}

func IPFromAddr(addr string) string {
	if len(addr) == 0 {
		return ""
	}
	taddr := addr
	if strings.HasPrefix(addr, def.TCP_PREFIX) {
		if addr == def.TCP_PREFIX {
			return ""
		}
		taddr = addr[len(def.TCP_PREFIX):]
	}
	index := strings.Index(taddr, ":")
	if index > 0 {
		return taddr[:index]
	}
	return ""
}

func CheckNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

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
	"fmt"
	"testing"
)

func TestOnlyNumLetterUnderline(t *testing.T) {
	str1 := "abc1abc2_"
	fmt.Println(str1, OnlyNumLetterUnderline(str1))
	str2 := "_12bc1_aAAAAVBBbc2"
	fmt.Println(str2, OnlyNumLetterUnderline(str2))
	str3 := "$" + str1
	fmt.Println(str3, OnlyNumLetterUnderline(str3))
	str4 := str1 + "_/\"" + str1
	fmt.Println(str4, OnlyNumLetterUnderline(str4))
	str5 := str1 + "_/\""
	fmt.Println(str5, OnlyNumLetterUnderline(str5))
	str6 := "%"
	fmt.Println(str6, OnlyNumLetterUnderline(str6))
}

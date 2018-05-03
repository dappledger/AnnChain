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

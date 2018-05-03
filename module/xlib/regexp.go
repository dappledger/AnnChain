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

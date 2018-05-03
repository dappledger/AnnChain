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

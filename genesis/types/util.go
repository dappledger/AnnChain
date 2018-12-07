package types

import (
	"strings"
)

func TrimZero(data []byte) []byte {
	if strings.HasPrefix(string(data), "0x") {
		data = data[2:]
	}
	for i, v := range data {
		if v != '0' {
			return data[i:]
		}
	}
	return data[:]
}

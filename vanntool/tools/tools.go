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

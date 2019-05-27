package utils

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_Crypto(t *testing.T) {
	text := "aa"
	pwd := "aa"
	ret, err := Encrypt([]byte(text), []byte(pwd))
	fmt.Println("Encrypt:", ret, ",err:", err, ",hex:", hex.EncodeToString(ret))
	deret, err := Decrypt(ret, []byte(pwd))
	fmt.Println("Decrypt:", string(deret), ",err:", err)
}

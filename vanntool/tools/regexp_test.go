package tools

import (
	"fmt"
	"net"
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

func TestCheckIP(t *testing.T) {
	ip1 := "0.0.0.0"
	fmt.Println(ip1, CheckIPAddr("tcp", ip1))
	ip2 := "0.0.0.0:1111"
	fmt.Println(ip2, CheckIPAddr("tcp", ip2))
	ip3 := "0.0.0.0:333_"
	fmt.Println(ip3, CheckIPAddr("tcp", ip3))
	ip4 := "tcp://" + ip2
	fmt.Println(ip4, CheckIPAddr("tcp", ip4))

	host, port, err := net.SplitHostPort(ip4)
	fmt.Printf("host:%v,port:%v,err:%v", host, port, err)
}

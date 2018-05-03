package node

import (
	"testing"
	"bufio"
	"bytes"
	"fmt"
	"io"
)

var data = "当输出数字的时候，你将经常想要控制输出结果的宽度和精度，可以使用在 % 后面使用数字来控制输出宽度。默认结果使用右对齐并且通过空格来填充空白部分。"

func TestCoSiWrap(t *testing.T) {
	
	wrapped := CosiWrapData( []byte(data))

	bfr := bufio.NewReader(bytes.NewReader(bytes.Join(wrapped, nil)))
	msgHeader, err := bfr.Peek(len(cosiHeader))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(msgHeader, cosiHeader) {
		t.Fatal("header is wrong")
	}

	bfr.Discard(len(cosiHeader))
	l, _ := bfr.Peek(8)
	if len([]byte(data)) != BytesToInt(l) {
		t.FailNow()
	}
	dataLen := BytesToInt(l)

	bfr.Discard(8)
	
	recvData := make([]byte, 0)

	buf := make([]byte, 50)
	for {
		n, err := bfr.Read(buf)
		if err != nil && err != io.EOF{
			t.Fatal(err)
		}
		recvData = append(recvData, buf[:n]...)
		if len(recvData) == dataLen {
			break
		}
	}

	fmt.Println(string(recvData))
}

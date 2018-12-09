package abi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	gin "gopkg.in/gin-gonic/gin.v1"
)

func HTTPGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func HttpJsonPost(url string, json []byte) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func WriteLine(fileName string, handler func(io.Writer)) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	handler(buf)
	buf.Flush()
	return nil
}

func ReadLine(fileName string, handler func(string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}

var (
	USR      = "public_info_mail@126.com"
	PWD      = "AbC123"
	HOST     = "smtp.126.com:25"
	TO_USERS = "fanbeishuang@126.com"
)

func SendToMail(user, password, host, to, title, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + title + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func FailSendMail(ctx *gin.Context, title, data string) {
	con, _ := ioutil.ReadAll(ctx.Request.Body)
	defer ctx.Request.Body.Close()

	body := fmt.Sprintf("quest_ip:%v,host:%v,url:%v,data:%v\npost:%v", ctx.ClientIP(), ctx.Request.Host, ctx.Request.URL.String(), data, string(con))

	if err := SendToMail(USR, PWD, HOST, TO_USERS, title, body, ""); err != nil {
		fmt.Println(err)
	}

}

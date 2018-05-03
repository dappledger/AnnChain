package ti

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	RandomCount = 8
	QueryFormat = "?key=%s&timestamp=%s&nonce=%s&version=%s&signature=%s"
	MacVersion  = "0.4.0"
)

type TiCapsuleClient struct {
	Endpoint    string
	UploadUrl   string
	DownloadUrl string
	Key         string
	Secret      string
	ApiClient   RestApiClient
}
type RestApiClient struct {
	Client                  http.Client
	connectionTimeoutMillis int
	socketTimeoutMillis     int
}
type UploadResult struct {
	IsSuccess bool   `json:"isSuccess"`
	Hash      string `json:"hash"`
	Info      string `json:"info"`
}

func NewTiCapsuleClient(endpoint, key, secret string) TiCapsuleClient {
	return TiCapsuleClient{
		Endpoint:    endpoint,
		Key:         key,
		Secret:      secret,
		UploadUrl:   fmt.Sprintf("%s/api/v0/save", endpoint),
		DownloadUrl: fmt.Sprintf("%s/api/v0/get/", endpoint),
	}
}
func (ti *TiCapsuleClient) Save(path string) (result UploadResult, err error) {
	query := NewHmacQuery(ti.Key, ti.Secret)
	bytez, err := ti.ApiClient.UploadFile(ti.UploadUrl+query, path)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytez, &result)
	return
}

func (ti *TiCapsuleClient) SaveData(name string, data []byte) (result UploadResult, err error) {
	query := NewHmacQuery(ti.Key, ti.Secret)
	bytez, err := ti.ApiClient.UploadData(ti.UploadUrl+query, name, data)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytez, &result)
	return
}

func (ti *TiCapsuleClient) DownloadFile(hash, path string) (err error) {
	query := NewHmacQuery(ti.Key, ti.Secret)
	url := ti.DownloadUrl + hash + query
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	file, err := os.Create(path)
	if err != nil {
		return
	}
	_, err = io.Copy(file, resp.Body)
	return
}

func (ti *TiCapsuleClient) DownloadContent(hash string) (content []byte, err error) {
	query := NewHmacQuery(ti.Key, ti.Secret)
	url := ti.DownloadUrl + hash + query
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	content = body
	return
}

func (cli *RestApiClient) UploadFile(url, path string) (body []byte, err error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("file", path)
	if err != nil {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(fw, file)
	if err != nil {
		return
	}
	w.Close()
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := cli.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("UploadFile response err: " + string(body))
	}
	return
}

func (cli *RestApiClient) UploadData(url, name string, data []byte) (body []byte, err error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("file", name)
	if err != nil {
		return
	}

	_, err = io.Copy(fw, bytes.NewBuffer(data))

	if err != nil {
		return
	}
	w.Close()
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := cli.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("UploadFile response err: " + string(body))
	}
	return
}

func NewHmacQuery(key, secret string) string {
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	nonce := creatNonce()
	data := key + timeStamp + nonce
	signature := sign(secret, data)
	return fmt.Sprintf(QueryFormat, key, timeStamp, nonce, MacVersion, signature)
}

func creatNonce() string {
	return RandomAlphanumeric(RandomCount)
}

func sign(secret, data string) string {
	return hmacSha1Hex(secret, data)
}
func hmacSha1Hex(secret, data string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

package util

import (
	"bytes"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// ClientInfo HTTP客户端信息
type ClientInfo struct {
	client        *http.Client      //请求实例
	urlParam      string            //url参数
	reqMethond    string            //http请求方法
	header        map[string]string //http 头
	reqTimeoutSec uint32            //请求超时时间s
	byteBuf       *bytes.Buffer
	multiWriter   *multipart.Writer
	isFormData    bool
}

func (clientInfo *ClientInfo) SetMethod(method string) {
	clientInfo.reqMethond = method
}
func (clientInfo *ClientInfo) SetTimeout(sec uint32) {
	clientInfo.reqTimeoutSec = sec
}
func (clientInfo *ClientInfo) SetContentType(value string) {
	if clientInfo.header == nil {
		log.Info("SetContentType fail,no init,call after InitHTTPClient")
		return
	}
	clientInfo.header["Content-Type"] = value
}
func (clientInfo *ClientInfo) AddHeader(key, value string) {
	if clientInfo.header == nil {
		log.Info("AddHeader fail,no init,call after InitHTTPClient")
		return
	}
	clientInfo.header[key] = value
}
func (clientInfo *ClientInfo) AddFormData(key, value string) {
	ioWriter, err := clientInfo.multiWriter.CreateFormField(key)
	if err != nil {
		log.Error("AddFormData CreateFormField fail", key, " value:", value)
		return
	}
	ioWriter.Write([]byte(value))
	clientInfo.isFormData = true
}
func (clientInfo *ClientInfo) AddFormFile(key, fileName, content string, isPath bool) bool {
	w, err := clientInfo.multiWriter.CreateFormFile(key, fileName)
	if err != nil {
		log.Error("add form file fail,CreateFormFile", err.Error())
		return false
	}
	if isPath {
		dataByte, err := ioutil.ReadFile(content)
		if err != nil {
			log.Error("add form file fail,", content, err.Error())
			return false
		}
		w.Write(dataByte)
	} else {
		w.Write([]byte(content))
	}
	clientInfo.isFormData = true
	return true
}

// InitHTTPClient 初始化HTTP
func (clientInfo *ClientInfo) InitHTTPClient() {
	clientInfo.header = make(map[string]string)
	clientInfo.reqMethond = "GET"
	clientInfo.reqTimeoutSec = 60
	clientInfo.SetContentType("application/json")
	clientInfo.byteBuf = new(bytes.Buffer)
	clientInfo.multiWriter = multipart.NewWriter(clientInfo.byteBuf)
}

// SetURLParam 设置URL参数
func (clientInfo *ClientInfo) SetURLParam(key string, value string) {
	if clientInfo.urlParam != "" {
		clientInfo.urlParam += "&" + key + "=" + value
	} else {
		clientInfo.urlParam += key + "=" + value
	}
}

func (clientInfo *ClientInfo) clearParam() {
	clientInfo.urlParam = ""
	clientInfo.reqMethond = ""
	clientInfo.header = make(map[string]string)
	if !clientInfo.isFormData {
		clientInfo.multiWriter.Close()
	}
}

// SendURLRequest 发送请求
func (clientInfo *ClientInfo) SendURLRequest(url, content string) string {
	defer ExceptionRecover()
	defer clientInfo.clearParam()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	clientInfo.client = &http.Client{
		Timeout:   time.Duration(clientInfo.reqTimeoutSec) * time.Second,
		Transport: tr}
	if clientInfo.client == nil {
		log.Error("SendURLRequest err,clientInfo.client is nil,forget call InitHTTPClient ?")
		return ""
	}
	//提交请求
	if clientInfo.urlParam != "" {
		url += "?" + clientInfo.urlParam
	}
	var (
		reqest *http.Request
		err    error
	)
	if clientInfo.isFormData {
		clientInfo.multiWriter.Close()
		clientInfo.reqMethond = "POST"
		reqest, err = http.NewRequest(clientInfo.reqMethond, url, clientInfo.byteBuf)
		delete(clientInfo.header, "Content-Type")
		reqest.Header.Set("Content-Type", clientInfo.multiWriter.FormDataContentType())

	} else {
		if content == "" {
			reqest, err = http.NewRequest(clientInfo.reqMethond, url, nil)
		} else {
			reqBody := strings.NewReader(content)
			reqest, err = http.NewRequest(clientInfo.reqMethond, url, reqBody)
		}
	}
	if nil != clientInfo.header {
		for key, value := range clientInfo.header {
			reqest.Header.Add(key, value)
		}
	}

	if err != nil {
		panic(err)
	}
	//处理返回结果
	resp, err := clientInfo.client.Do(reqest)
	if err != nil {
		log.Error(fmt.Sprintf("SendURLRequest err,call Do,err msg:%s", err.Error()))
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("SendURLRequest err,call ioutil.ReadAll,err msg:%s", err.Error()))
	}
	return string(body)
}

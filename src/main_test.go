package main

import (
	_ "fmt"
	"testing"
)

func TestInitParam(t *testing.T) {
	rp := RequestParam{}
	rp.init()
	if rp.clients != 1 {
		t.Error("Init Error, check clients")
		t.Fail()
		return
	}
	if rp.ua != PROGRAM_NAME+PROGRAM_VERSION {
		t.Error("Init Error, check ua")
		t.Fail()
		return
	}
	if rp.method != "GET" {
		t.Error("Init Error, check method")
		t.Fail()
		return
	}
	if rp.defaultTime != 30 {
		t.Error("Init Error, check running time")
		t.Fail()
		return
	}
	if rp.verbose != false {
		t.Error("Init Error, check is verbose")
		t.Fail()
		return
	}
	if rp.url != "" {
		t.Error("Init Error, check url")
		t.Fail()
		return
	}
	t.Log("\nRequest Param:\n", rp)
	t.Log("PASS: Init Param")
}

func TestHTTP(t *testing.T) {
	// send HTTP
	//buildRequest("http://www.baidu.com", "GET")
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "http://www.baidu.com"
	requestParam.defaultTime = 1
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan string)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		//t.Log(i)
		i += ""
		tr += 1
	}
	t.Log("Send: ", tr, " times requests")
	t.Log("PASS: HTTP request")
}

func TestHTTPs(t *testing.T) {
	// send HTTPS
	//buildRequest("https://www.baidu.com", "GET")
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "https://www.baidu.com"
	requestParam.defaultTime = 1
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan string)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		//t.Log(i)
		i += ""
		tr += 1
	}
	t.Log("Send: ", tr, " times requests")
	t.Log("PASS: HTTPs request")
}

func TestHTTP2(t *testing.T) {
	t.Log("Not support Force HTTP2 yet")
	t.Fail()
}
func TestMultiClient(t *testing.T) {
	// -c param
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "http://www.baidu.com"
	requestParam.defaultTime = 1
	requestParam.clients = 10
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan string)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		//t.Log(i)
		i += ""
		tr += 1
	}
	t.Log("Send: ", tr, " times requests")
	t.Log("PASS: Multi client")
}

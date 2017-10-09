package main

import (
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
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "http://www.baidu.com"
	requestParam.proto = "HTTP/1.1"
	requestParam.defaultTime = 1
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan HttpRes)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		if i.errNo != 0 {
			t.Error("FAILED: HTTP request")
			t.Fail()
			return
		}
		//t.Log(i.errNo)
		tr += 1
	}
	t.Log("Send: ", tr, " times requests")
	t.Log("PASS: HTTP request")
}

func TestHTTPs(t *testing.T) {
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "https://www.baidu.com"
	requestParam.defaultTime = 1
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan HttpRes)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		if i.errNo != 0 {
			t.Error("FAILED: HTTPs request")
			t.Fail()
			return
		}
		//t.Log(i)
		tr += 1
	}
	t.Log("Send: ", tr, " times requests")
	t.Log("PASS: HTTPs request")
}

func TestHTTP2(t *testing.T) {
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "https://http2.golang.org/reqinfo"
	requestParam.defaultTime = 1
	requestParam.proto = "HTTP/2.0"
	requestParam.protoMajor = 2
	requestParam.protoMinor = 0
	requestParam.tr = nil
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan HttpRes)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		if i.errNo != 0 {
			t.Error("FAILED: HTTP/2 request")
			t.Fail()
			return
		}
		//t.Log(i)
		tr += 1
	}
	t.Log("PASS: HTTP/2 request")
}

func TestMultiClient(t *testing.T) {
	// -c param
	var requestParam RequestParam
	requestParam.init()
	requestParam.url = "http://www.baidu.com"
	requestParam.defaultTime = 1
	requestParam.clients = 10
	t.Log("\nRequest Param:\n", requestParam)
	ch := make(chan HttpRes)
	cch := make(chan int, 1)
	go sendHTTPRequest(requestParam, ch, cch, 0)
	tr := 0
	for i := range ch {
		if i.errNo != 0 {
			t.Error("FAILED: MultiClient request")
			t.Fail()
			return
		}
		//t.Log(i)
		tr += 1
	}
	t.Log("Send: ", tr, " times requests")
	t.Log("PASS: Multi client")
}

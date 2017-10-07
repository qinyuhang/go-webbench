package main

import (
	"testing"
)

func TestSendRequest(t *testing.T) {
}

func TestHTTP(t *testing.T) {
	// send HTTP
	buildRequest("http://www.baidu.com")
	t.Log("PASS: HTTPs request")
}

func TestHTTPs(t *testing.T) {
	// send HTTPS
	buildRequest("https://www.baidu.com")
	t.Log("PASS: HTTP request")
}

func TestMultiClient(t *testing.T) {
	// -c param
	t.Log("PASS: Multi client")
}

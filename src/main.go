//
// (C) qinyuhangxiaoxiang@gmail.com 2017
// This is free software, see GNU Public License version 3 for
// details.
//
// Webbench based on go language
// mainly same with Webbench on C (see https://github.com/qinyuhang/WebBench/)
//
// Add support for HTTPs
// And will support HTTP/2
//
// Usage:
//   webbench -h
//
//
//
package main

import (
	"bytes"
	"crypto/tls"
	_ "encoding/json"
	_ "encoding/xml"
	"fmt"
	_ "io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	_ "strings"
	_ "sync"
	"time"
)

var PROGRAM_VERSION string = "0.1"
var PROGRAM_NAME string = "GoWebbench"

// HTTPRes ...
// this Data type is suppose to be the data type that
// all sendHttpRequest func put in first ch
// main func take out of the value
// and see if the is a error or not
// if error stop the program and display error
// if not show the showText in console
type HTTPRes struct {
	// if -v is provided, print this text to console
	showText string

	// resCode is response.StatusCode
	resCode int

	// resp is pointer to the response of the request
	resp *http.Response

	// errNo could be 0, 1
	// if 0 no error
	// if 1 has error detail see errMsg
	errNo int

	// if no error errMsg = ""
	// or errMsg show what error is
	errMsg string
}

func (h *HTTPRes) init(showText string, resp *http.Response, errNo int, errMsg string) {
	h.showText = showText
	if resp != nil {
		h.resCode = resp.StatusCode
	} else {
		h.resCode = 200
	}
	h.resp = resp
	h.errNo = errNo
	h.errMsg = errMsg
}

// RequestParam ...
// the data struct help build HTTP request
type RequestParam struct {
	// How many clients running
	clients int

	// request Method
	method string

	// how long does the request running
	defaultTime int

	// should show verbose info in console?
	verbose bool

	// request url
	url string

	// request protocol now support HTTP/1.1 HTTP/2
	proto string

	protoMajor int

	protoMinor int

	// http Transport object
	tr *http.Transport

	// all HTTP headers
	headers map[string]string

	// all body, will ignore when send GET HEAD OPTIONS
	body []byte

	// not read res body
	force bool
}

func (requestParam *RequestParam) init() {
	// init request params
	requestParam.clients = 1
	requestParam.method = "GET"
	requestParam.defaultTime = 30
	requestParam.verbose = false
	requestParam.url = ""
	// default protocol is HTTP/1.1
	requestParam.proto = "HTTP/1.1"
	requestParam.protoMajor = 1
	requestParam.protoMinor = 1
	// default HTTP/1.1 need disable HTTP/2.0 by set this
	requestParam.tr = &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	requestParam.headers = make(map[string]string)
	requestParam.headers["User-Agent"] = fmt.Sprint(PROGRAM_NAME, " version ", PROGRAM_VERSION)
	requestParam.body = nil
	requestParam.force = false
}

func (requestParam RequestParam) String() string {
	return fmt.Sprint(
		"clients: ", requestParam.clients, "\n",
		"url: ", requestParam.url, "\n",
		"method: ", requestParam.method, "\n",
		"protocol: ", requestParam.proto, "\n",
		"isVerbose: ", requestParam.verbose, "\n",
		"runningTime: ", requestParam.defaultTime, "\n",
		"headers: ", requestParam.headers, "\n",
		"isForce: ", requestParam.force, "\n",
	)
}

func initUAMap(userAgentMap *map[string]string) {
	*userAgentMap = map[string]string{
		"iOSWechat":     "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 9_2 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13C75 MicroMessager/6.3.9",
		"AndroidWechat": "User-Agent: Mozilla/5.0 (Linux; Android 4.4.2; HUAWEI G750-T20 Build/HuaweiG750-T20) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile MicroMessager/6.3.9 NetType/WIFI Language/zh_CN",
		"iOS9_2":        "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 9_2 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13C75",
		"Android":       "User-Agent: Mozilla/5.0 (Linux; Android 4.4.2; HUAWEI G750-T20 Build/HuaweiG750-T20) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile Safari/537.36",
	}
}

func initArgsMap(sam *map[string]func(c string) int, requestParam *RequestParam, userAgentMap map[string]string) {
	*sam = map[string]func(c string) int{
		"-c": func(c string) int {
			// clients
			if res, ok := strconv.Atoi(c); ok == nil {
				requestParam.clients = res
			}
			return 1
		},
		"-d": func(c string) int {
			return 1
		},
		"-u": func(c string) int {
			requestParam.headers["User-Agent"] = userAgentMap[c]
			return 1
		},
		"-V": func(c string) int {
			showProgramVersion()
			return 0
		},
		"-1": func(c string) int {
			// HTTP1
			requestParam.proto = "HTTP/1.0"
			requestParam.protoMajor = 1
			requestParam.protoMinor = 0
			// disable HTTP2
			requestParam.tr = &http.Transport{
				DisableKeepAlives: false,
				TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			}
			return 0
		},
		"-1.1": func(c string) int {
			// HTTP 1.1
			requestParam.proto = "HTTP/1.1"
			requestParam.protoMajor = 1
			requestParam.protoMinor = 1
			// disable HTTP/2
			requestParam.tr = &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			}
			requestParam.headers["Connection"] = "close"
			return 0
		},
		"-2": func(c string) int {
			// HTTP2
			requestParam.proto = "HTTP/2.0"
			requestParam.protoMajor = 2
			requestParam.protoMinor = 0
			requestParam.tr = nil
			return 0
		},
		"-t": func(c string) int {
			t, err := strconv.Atoi(c)
			if err == nil {
				requestParam.defaultTime = t
			}
			return 1
		},
		"--get": func(c string) int {
			requestParam.method = "GET"
			return 0
		},
		"--post": func(c string) int {
			requestParam.method = "POST"
			return 0
		},
		"--head": func(c string) int {
			requestParam.method = "HEAD"
			return 0
		},
		"--options": func(c string) int {
			requestParam.method = "OPTIONS"
			return 0
		},
		"--delete": func(c string) int {
			requestParam.method = "DELETE"
			return 0
		},
		"--put": func(c string) int {
			requestParam.method = "PUT"
			return 0
		},
		"--trace": func(c string) int {
			requestParam.method = "TRACE"
			return 0
		},
		"--connect": func(c string) int {
			requestParam.method = "CONNECT"
			return 0
		},
		"--patch": func(c string) int {
			requestParam.method = "PATCH"
			return 0
		},
		"-v": func(c string) int {
			requestParam.verbose = true
			return 0
		},
		"-h": func(c string) int {
			usage()
			return 0
		},
		"-?": func(c string) int {
			usage()
			return 0
		},
		"-r": func(c string) int {
			requestParam.headers["Pragma"] = "no-cache"
			return 0
		},
		"-f": func(c string) int {
			requestParam.force = true
			return 0
		},
		"-F": func(c string) int {
			// c is a string of JSON
			fmt.Println(c)
			jsonSlice := []byte(c)
			fmt.Println(jsonSlice)
			requestParam.body = jsonSlice
			//dec := json.NewDecoder(strings.NewReader(c))
			//for {
			//	t, err := dec.Token()
			//	if err == io.EOF {
			//		break
			//	}
			//	if err != nil {
			//		log.Fatal(err)
			//	}
			//	fmt.Printf("%T: %v", t, t)
			//	if fmt.Sprintf("%T", t) != "json.Delim" && fmt.Sprintf("%v", t) == "{" {
			//		// Start of a json Object could be the very first or some value
			//
			//	}
			//	if dec.More() {
			//		fmt.Printf(" (more)")
			//	}
			//	fmt.Printf("\n")
			//}
			return 1
		},
	}
}

func main() {
	var requestParam RequestParam
	requestParam.init()
	//fmt.Println(requestParam)

	// program consts
	var coordinateCh chan int
	var ch chan HTTPRes
	var argsCount int
	var usefulArgsCount int
	var totalRequestCount uint64
	var totalRequestSize int64

	argsMap := make(map[string]string)

	userAgentMap := make(map[string]string)
	initUAMap(&userAgentMap)

	var supportArgsMap map[string]func(c string) int
	initArgsMap(&supportArgsMap, &requestParam, userAgentMap)

	// main program
	argsCount = len(os.Args)
	usefulArgsCount = 0
	//fmt.Println(userAgentMap["iOSWechat"])
	// Starting the main program
	fmt.Println("GoWebbanch versioin", PROGRAM_VERSION)
	for i := 0; i < argsCount; i += 1 {
		v := os.Args[i]
		//fmt.Println(i, v)
		if supportArgsMap[v] != nil {
			usefulArgsCount += 1
			if i+1 < argsCount && []byte(os.Args[i+1])[0] != []byte("-")[0] {
				// has next param and next param is not start with "-"
				if supportArgsMap[v](os.Args[i+1]) == 1 {
					// only when ret == 1 means the next param is taken
					argsMap[v] = os.Args[i+1]
					i += 1
				}
			} else {
				// --options .etc will go here
				supportArgsMap[v]("")
			}
		} else if i != 0 {
			// i != 0 去除 os.Args 中程序名 等等的干扰
			re := regexp.MustCompile("^http://")
			re1 := regexp.MustCompile("^https://")
			if re.Find([]byte(v)) != nil || re1.Find([]byte(v)) != nil {
				// Have url
				usefulArgsCount += 1
				requestParam.url = v
			} else {
				//no url
				fmt.Println("GoWebbench missing url!")
				usage()
				return
			}
		}
	}
	if usefulArgsCount == 0 {
		usage()
		return
	}
	// init all channels
	ch = make(chan HTTPRes)
	coordinateCh = make(chan int, requestParam.clients)
	//failCh := make(chan string)
	fmt.Println(requestParam)
	if requestParam.clients != 0 {
		for i := 0; i < requestParam.clients; i += 1 {
			go sendHTTPRequest(requestParam, ch, coordinateCh, i)
		}
	}
	totalRequestCount = 0
	totalRequestSize = 0
	for i := range ch {
		if requestParam.verbose {
			fmt.Println(i)
		}
		if i.errNo != 0 {
			fmt.Print("\n", i.errMsg, "\n")
			return
		}
		if i.resp.ContentLength != -1 {
			totalRequestSize += i.resp.ContentLength
		}
		totalRequestCount += 1
	}
	fmt.Println("Total request", totalRequestCount, "in", requestParam.defaultTime, "s")
	fmt.Println("Speed is", totalRequestCount/uint64(requestParam.defaultTime), "pages/sec")
	fmt.Println("Speed is", totalRequestCount/uint64(requestParam.defaultTime)*60, "pages/min")
	fmt.Println("Speed is", totalRequestSize/int64(requestParam.defaultTime), "bytes/sec")
}

func httpCodeHandler(httpCode int) {

}

func sendHTTPRequest(requestParam RequestParam, ch chan<- HTTPRes, coordinateCh chan int, label int) int {
	// should put a struct that contain the res time and result in the res channel
	requestDuration, reqDErr := time.ParseDuration(strconv.Itoa(requestParam.defaultTime) + "s")
	if reqDErr == nil {
		//fmt.Println(requestDuration)
		//timeout := time.NewTimer(requestDuration)
		routineStartTime := time.Now()
		var client *http.Client
		if requestParam.tr != nil {
			client = &http.Client{
				Timeout:   time.Second * 10,
				Transport: requestParam.tr,
			}
		} else {
			client = &http.Client{
				Timeout: time.Second * 10,
			}
		}
	Loop:
		for {
			if time.Since(routineStartTime) < requestDuration {
				requestStartTime := time.Now()
				reqBody := buildRequest(requestParam)

				resp, err := client.Do(reqBody)

				var resData HTTPRes
				if err != nil {
					// push a failed state to a counter
					// TODO not only put this in but alsof signal the main goroutine it has a error
					resData.init(
						fmt.Sprintf("%s%s Label %d", "Error ", err, label),
						resp,
						1,
						fmt.Sprintf("%s%s Label %d", "Error ", err, label),
					)
					ch <- resData
				} else {
					//fmt.Println(resp.Body)
					//fmt.Println(resp.StatusCode)
					//fmt.Println(resp.ContentLength)
					// TODO if HTTP Proto version is given and res version not match should fail
					// TODO what failed DO? send Error struct in ch and send label to coordinateCh
					// TODO and if the coordinateCh if full close above 2 channels
					//fmt.Println(resp.Proto)
					//fmt.Println(resp.Request.Proto)
					//fmt.Println(resp.Request.Proto == resp.Proto)
					errNo := 0
					errMsg := ""
					if resp.Request.Proto != resp.Proto {
						errNo = 1
						errMsg = fmt.Sprint("Failed! Response Protocol Not match, want ", resp.Request.Proto, " get ", resp.Proto)
					}
					if requestParam.force {
						// while have -f we want not spend time read it
						go func() {
							ioutil.ReadAll(resp.Body)
							resp.Body.Close()
						}()
					} else {
						respBody, err := ioutil.ReadAll(resp.Body)
						resp.Body.Close()
						if err != nil {
							log.Fatal(err)
						}
						if resp.ContentLength == -1 {
							resp.ContentLength = int64(len(respBody))
						}
					}
					//fmt.Printf("%s", robots)
					endTime := time.Since(requestStartTime).Seconds()
					resData.init(
						fmt.Sprint("Client No.", label, ",Fetch url ", requestParam.url, " ,runing ", endTime, "s"),
						resp,
						errNo,
						errMsg,
					)
					ch <- resData
				}
			} else {
				coordinateCh <- label
				if cap(coordinateCh) == len(coordinateCh) {
					// 本 goroutine 放到coordinateCh 之后如果协作channel已满，则表示本goroutine是最后一个完成的
					for i := 0; i < len(coordinateCh); i += 1 {
						// 取出所有协作channel的值
						<-coordinateCh
					}
					close(coordinateCh)
					close(ch)
				}
				break Loop
			}
		}
	}
	return 1
}

func buildRequest(requestParam RequestParam) *http.Request {
	// maybe should change the requestBody from string to *http.request
	req, err := http.NewRequest(requestParam.method, requestParam.url, bytes.NewBuffer(requestParam.body))
	//fmt.Println(method)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
		return nil
	}
	req.Proto = requestParam.proto
	req.ProtoMajor = requestParam.protoMajor
	req.ProtoMinor = requestParam.protoMinor
	for i, v := range requestParam.headers {
		req.Header.Add(i, v)
	}
	return req
}

func usage() {
	x := fmt.Sprint("webbench [option]... URL\n",
		"  -f  |  --force               Don't wait for reply from server.\n",
		"  -r  |  --reload              Send reload request - Pragma: no-cache.\n",
		"  -d  |  --data                Read POST body from csv or json file.\n",
		"  -F  |  --Field               Read the Field added to request body from csv or json file.\n",
		"  -i  |  --input				Read the file Path json file as config",
		"  -t  |  --time <sec>          Run benchmark for <sec> seconds. Default 30.\n",
		"  -p  |  --proxy <server:port> Use proxy server for request.\n",
		"  -c  |  --clients <n>         Run <n> HTTP clients at once. Default one.\n",
		"  -0.9|  --http0.9           Use HTTP/0.9 style requests.\n",
		"  -1  |  --http1.0             Use HTTP/0.9 style requests.\n",
		"  -1.1|  --http1.1           Use HTTP/1.0 protocol.\n",
		"  -2  |  --http2               Use HTTP/1.1 protocol.\n",
		"  -u  |  --User-Agent          Change User-Agent.\n",
		"  \t 1 for WeCaht iPhone\t 2 for WeChat Android\t 3 for iPhone Safari\n",
		"  \t 4 for Android Chrome\t 5 for Windows IE11\t 6 for Windows IE10\n",
		"  \t 7 for Windows Edge\t 8 for Windows Chrome\t 9 for Windows FireFox\n",
		"  \t 10 for Mac Safari\t 11 for Mac Chrome\t 12 for Mac FireFox\n",
		"  Empty for WebBench Program Version.\n",
		"  --get                    Use GET request method.\n",
		"  --head                   Use HEAD request method.\n",
		"  --options                Use OPTIONS request method.\n",
		"  --trace                  Use TRACE request method.\n",
		"  --post                   Use POST request method.\n",
		"  --delet                  Use DELETE request method.\n",
		"  --put                    Use PUT request method.\n",
		"  --connect                Use CONNECT request method.\n",
		"  --patch                  Use PATCH request method.\n",
		"  -?  |  -h|--help             This information.\n",
		"  -V  |  --version             Display program version.\n",
		"  -v  |  --verbose             Show Every Request.\n")
	fmt.Println(x)
}

func showProgramVersion() {
	fmt.Println("GoWebbench versoin ", PROGRAM_VERSION)
}

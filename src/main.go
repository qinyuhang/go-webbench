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
	"crypto/tls"
	_ "encoding/json"
	_ "encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	_ "sync"
	"time"
)

var PROGRAM_VERSION string = "0.1"
var PROGRAM_NAME string = "GoWebbench"

type HttpRes struct {
	showText string
	resCode  int
	resp     *http.Response
}

func (h *HttpRes) init(showText string, resp *http.Response) {
	h.showText = showText
	h.resCode = resp.StatusCode
	h.resp = resp
}

type RequestParam struct {
	clients     int
	ua          string
	method      string
	defaultTime int
	verbose     bool
	url         string
	proto       string
	tr          *http.Transport
}

func (requestParam *RequestParam) init() {
	// init request params
	requestParam.clients = 1
	requestParam.ua = PROGRAM_NAME + PROGRAM_VERSION
	requestParam.method = "GET"
	requestParam.defaultTime = 30
	requestParam.verbose = false
	requestParam.url = ""
	requestParam.proto = "HTTP/2"
	requestParam.tr = nil
}

func (requestParam RequestParam) String() string {
	return fmt.Sprint(
		"clients: ", requestParam.clients, "\n",
		"ua: ", requestParam.ua, "\n",
		"url: ", requestParam.url, "\n",
		"method: ", requestParam.method, "\n",
		"isVerbose: ", requestParam.verbose, "\n",
		"runningTime: ", requestParam.defaultTime, "\n",
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
			requestParam.ua = userAgentMap[c]
			return 1
		},
		"-V": func(c string) int {
			showProgramVersion()
			return 0
		},
		"-1": func(c string) int {
			// HTTP1
			requestParam.proto = "HTTP/1.0"
			// disable HTTP2
			requestParam.tr = &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			}
			return 0
		},
		"-1.1": func(c string) int {
			// HTTP 1.1
			requestParam.proto = "HTTP/1.1"
			// disable HTTP/2
			requestParam.tr = &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			}
			return 0
		},
		"-2": func(c string) int {
			// HTTP2
			requestParam.proto = "HTTP/2"
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
	}
}

func main() {
	var requestParam RequestParam
	requestParam.init()
	//fmt.Println(requestParam)

	// program consts
	var coordinateCh chan int
	var ch chan string
	var argsCount int
	var usefulArgsCount int
	var totalRequestCount uint64

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
	fmt.Println("GoWebbanch versioin ", PROGRAM_VERSION)
	for i := 0; i < argsCount; i += 1 {
		v := os.Args[i]
		//fmt.Println(i, v)
		if supportArgsMap[v] != nil {
			usefulArgsCount += 1
			if i+1 < argsCount && []byte(os.Args[i+1])[0] != []byte("-")[0] {
				// has next param and next param is not start with "-"
				if supportArgsMap[v](os.Args[i+1]) == 1 {
					argsMap[v] = os.Args[i+1]
					// only when ret == 1 means the next param is taken
					i += 1
				}
				//fmt.Println(v, " is got", os.Args[i+1])
				//fmt.Println(ret)
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
	ch = make(chan string)
	coordinateCh = make(chan int, requestParam.clients)
	//failCh := make(chan string)
	if requestParam.clients != 0 {
		//fmt.Println(requestParam)
		for i := 0; i < requestParam.clients; i += 1 {
			go sendHTTPRequest(requestParam, ch, coordinateCh, i)
		}
	}
	//for i := 0; i < requestParam.clients; i += 1 {
	//	fmt.Println(<-ch)
	//}
	totalRequestCount = 0
	for i := range ch {
		if requestParam.verbose {
			fmt.Println(i)
		}
		totalRequestCount += 1
	}
	fmt.Println("Total request", totalRequestCount, "in", requestParam.defaultTime, "s")
	fmt.Println("Speed is", totalRequestCount/uint64(requestParam.defaultTime), "pages per second")
}

func httpCodeHandler(httpCode int) {

}

func sendHTTPRequest(requestParam RequestParam, ch chan<- string, coordinateCh chan int, label int) int {
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

				if err != nil {
					// push a failed state to a counter
					// TODO not only put this in but alsof signal the main goroutine it has a error
					ch <- fmt.Sprintf("%s%s Label %s", "Error ", err, label)
				} else {
					//fmt.Println(resp.Body)
					//fmt.Println(resp.StatusCode)
					//fmt.Println(resp.ContentLength)
					// TODO if HTTP Proto version is given and res version not match should fail
					//fmt.Println(resp.Proto)
					_, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
						log.Fatal(err)
					}
					//fmt.Printf("%s", robots)
					endTime := time.Since(requestStartTime).Seconds()
					ret := fmt.Sprint("Client No.", label, ",Fetch url ", requestParam.url, ",runing ", endTime, "s")
					ch <- ret
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
	req, err := http.NewRequest(requestParam.method, requestParam.url, nil)
	//fmt.Println(method)
	if err != nil {
		return nil
	}
	req.Proto = requestParam.proto
	req.Header.Add("User-Agent", requestParam.ua)
	return req
}

func usage() {
	x := fmt.Sprint("webbench [option]... URL\n",
		"  -f|--force               Don't wait for reply from server.\n",
		"  -r|--reload              Send reload request - Pragma: no-cache.\n",
		"  -d|--data                Read POST body from csv or json file.\n",
		"  -F|--Field               Read the Field added to request body from csv or json file.\n",
		"  -i|--input				Read the file Path json file as config",
		"  -t|--time <sec>          Run benchmark for <sec> seconds. Default 30.\n",
		"  -p|--proxy <server:port> Use proxy server for request.\n",
		"  -c|--clients <n>         Run <n> HTTP clients at once. Default one.\n",
		"  -9|--http09              Use HTTP/0.9 style requests.\n",
		"  -1|--http10              Use HTTP/1.0 protocol.\n",
		"  -2|--http11              Use HTTP/1.1 protocol.\n",
		"  -u|--User-Agent          Change User-Agent.\n",
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
		"  -?|-h|--help             This information.\n",
		"  -V|--version             Display program version.\n",
		"  -v|--verbose             Show Every Request.\n")
	fmt.Println(x)
}

func showProgramVersion() {
	fmt.Println("GoWebbench versoin ", PROGRAM_VERSION)
}

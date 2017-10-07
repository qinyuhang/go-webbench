// webbench based on go language
// mainly with webbench on C
// add support for HTTPS
// and add support for HTTP/2
package main

import (
	_ "encoding/json"
	_ "encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var PROGRAM_VERSION string = "0.1"

func main() {
	// request params
	var clients int = 1
	var ua string

	// program consts
	type safeCounter struct {
		number int
		m      sync.Mutex
	}
	var syncCh chan int
	argsMap := make(map[string]string)

	userAgentMap := map[string]string{
		"iOSWechat":     "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 9_2 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13C75 MicroMessager/6.3.9",
		"AndroidWechat": "User-Agent: Mozilla/5.0 (Linux; Android 4.4.2; HUAWEI G750-T20 Build/HuaweiG750-T20) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile MicroMessager/6.3.9 NetType/WIFI Language/zh_CN",
		"iOS9_2":        "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 9_2 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Mobile/13C75",
		"Android":       "User-Agent: Mozilla/5.0 (Linux; Android 4.4.2; HUAWEI G750-T20 Build/HuaweiG750-T20) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile Safari/537.36",
	}

	supportArgsMap := map[string]func(c string) int{
		"-c": func(c string) int {
			// clients
			res, ok := strconv.Atoi(c)
			if ok != nil {
				clients = 0
				syncCh = nil
			} else {
				clients = res
				syncCh = make(chan int, res)
			}
			return 1
		},
		"-d": func(c string) int {
			return 1
		},
		"-u": func(c string) int {
			ua = userAgentMap[c]
			return 1
		},
		"-V": func(c string) int {
			showProgramVersion()
			return 1
		},
	}

	// main program
	arg_num := len(os.Args)
	total_num := 0
	requestURL := ""
	//fmt.Println(userAgentMap["iOSWechat"])
	// Starting the main program
	fmt.Println("GoWebbanch versioin ", PROGRAM_VERSION)
	for i := 0; i < arg_num; i += 1 {
		v := os.Args[i]
		//fmt.Println(i, v)
		if supportArgsMap[v] != nil {
			total_num += 1
			// TODO judge should the -x have a extra param
			if i+1 < arg_num && []byte(os.Args[i+1])[0] != []byte("-")[0] {
				argsMap[v] = os.Args[i+1]
				supportArgsMap[v](os.Args[i+1])
				i += 1
				//fmt.Println(v, " is got", os.Args[i+1])
				//fmt.Println(ret)
			}
		} else if i != 0 {
			// i != 0 去除 os.Args 中程序名 等等的干扰
			re := regexp.MustCompile("^http://")
			re1 := regexp.MustCompile("^https://")
			if re.Find([]byte(v)) != nil || re1.Find([]byte(v)) != nil {
				// Have url
				total_num += 1
				requestURL = v
			} else {
				//no url
				fmt.Println("GoWebbench missing url!")
				usage()
				return
			}
		}
	}
	if total_num == 0 {
		usage()
		return
	}
	ch := make(chan string)
	//failCh := make(chan string)
	if clients != 0 {
		for i := 0; i < clients; i += 1 {
			go sendHTTPRequest(requestURL, ua, ch)
		}
	}
	for i := 0; i < clients; i += 1 {
		fmt.Println(<-ch)
	}
}

func sendHTTPRequest(url string, header string, ch chan<- string) int {
	// should return a struct that contain the res time and result
	startTime := time.Now()
	client := &http.Client{}
	reqBody := buildRequest(url)

	resp, err := client.Do(reqBody)

	if err != nil {
		// push a failed state to a counter
		ch <- fmt.Sprintf("%s%s", "Error ", err)
	} else {
		//fmt.Println(resp.Body)
		_, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s", robots)
		endTime := time.Since(startTime).Seconds()
		ret := fmt.Sprint("Fetch url ", url, " runing ", endTime, "s")
		ch <- ret
	}
	//fmt.Println(ret)
	return 1
}

func buildRequest(url string) *http.Request {
	// maybe should change the requestBody from string to *http.request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Add("User-Agent", fmt.Sprint("GoWebbanch versioin ", PROGRAM_VERSION))
	return req
}

func usage() {
	x := fmt.Sprint("webbench [option]... URL\n",
		"  -f|--force               Don't wait for reply from server.\n",
		"  -r|--reload              Send reload request - Pragma: no-cache.\n",
		"  -d|--data                Read POST body from csv or json file.\n",
		"  -F|--Field               Read the Field added to request body from csv or json file.\n",
		"  -t|--time <sec>          Run benchmark for <sec> seconds. Default 30.\n",
		"  -p|--proxy <server:port> Use proxy server for request.\n",
		"  -c|--clients <n>         Run <n> HTTP clients at once. Default one.\n",
		"  -9|--http09              Use HTTP/0.9 style requests.\n",
		"  -1|--http10              Use HTTP/1.0 protocol.\n",
		"  -2|--http11              Use HTTP/1.1 protocol.\n",
		"  -u|--User-Agent          Change User-Agent.\n",
		"  1 for WeCaht iPhone\t 2 for WeChat Android\t 3 for iPhone Safari\n",
		"  4 for Android Chrome\t 5 for Windows IE11\t 6 for Windows IE10\n",
		"  7 for Windows Edge\t 8 for Windows Chrome\t 9 for Windows FireFox\n",
		"  10 for Mac Safari\t 11 for Mac Chrome\t 12 for Mac FireFox\n",
		"  Empty for WebBench Program Version.\n",
		"  --get                    Use GET request method.\n",
		"  --head                   Use HEAD request method.\n",
		"  --options                Use OPTIONS request method.\n",
		"  --trace                  Use TRACE request method.\n",
		"  --post                   Use TRACE request method.\n",
		"  --delet                  Use TRACE request method.\n",
		"  --put                    Use TRACE request method.\n",
		"  --connect                Use TRACE request method.\n",
		"  -?|-h|--help             This information.\n",
		"  -V|--version             Display program version.\n")
	fmt.Println(x)
}

func showProgramVersion() {
	fmt.Println("GoWebbench versoin ", PROGRAM_VERSION)
}

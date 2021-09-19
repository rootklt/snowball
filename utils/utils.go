package utils

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/color"
)

var WarnOutput *color.Color = color.New(color.FgRed, color.Bold)
var SuccessOut *color.Color = color.New(color.FgGreen, color.Bold)

//输入地址，生成带协议的url，如127.0.0.1:8080 => http://127.0.0.1:8080
func GenUrl(u string) string {
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		if strings.HasSuffix(u, "443") {
			u = "https://" + u
		} else {
			u = "http://" + u
		}
	}
	return u
}

var ReqClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			DualStack: true,
		}).DialContext,
		Proxy: nil,
	},
	Timeout: 5 * time.Second,
}

func DoRequest(request *http.Request) (*http.Response, error) {

	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3100.0 Safari/537.36")

	response, err := ReqClient.Do(request)

	return response, err
}

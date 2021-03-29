package xray

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"snowball/utils"
	"time"

	"github.com/wonderivan/logger"
	"gopkg.in/yaml.v2"
)

//xray扫描到有漏洞时的请求数据
type ScanResults struct {
	Type string `json:"type" yaml:"type"`
	Data struct {
		Plugin     string `json:"plugin" yaml:"plugin"`
		CreateTime uint   `json:"create_time" yaml:"create_time"`
		Detail     struct {
			Addr     string                 `json:"addr" yaml:"addr"`
			Payload  string                 `json:"payload" yaml:"payload"`
			SnapShot []interface{}          `json:"snapshot" yaml:"snapshot"`
			Extra    map[string]interface{} `json:"extra" yaml:"extra"`
		}
	}
}

//xray统计信息请求数据
type WebHookCountor struct {
	Type string `json:"type"`
	Data struct {
		NumFoundUrls            int     `json:"num_found_urls"`
		NumScannedUls           int     `json:"num_scanned_urls"`
		NumSentHttpRequests     int     `json:"num_sent_http_requests"`
		AverageResponseTime     float32 `json:"average_response_time"`
		RadioFailedHttpRequests float32 `json:"ratio_failed_http_requests"`
	}
}

/*
//响应xray请求
type ResponseData struct {
	Status int    `json:"status"`
	Error  bool   `json:"error"`
	Msg    string `json:"msg"`
}
*/

//请求数据的类型，当请求的数据类型web_statistic时为web统计信息，为web_vuln时为漏洞信息，根据请求类型判断不同的数据
type RequestBodyType struct {
	Type string `json:"type" yaml:"type"`
}

type WebhookConfig struct {
	Webhook
}
type Webhook struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	Path string `yaml:"path"`
}

func (wf *WebhookConfig) ReadConfigFile() {
	const fileName = "config/config.yaml"
	buffer, err := ioutil.ReadFile(fileName)

	if err != nil {
		log.Printf("%v", err)
	}
	yaml.Unmarshal(buffer, wf)
}

func XrayWebhook(w http.ResponseWriter, r *http.Request) {
	const logfile = "log/vuls.log"

	rt := &RequestBodyType{}

	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error")
	}

	json.Unmarshal(result, rt)
	if rt.Type == "web_vuln" {
		res := &ScanResults{}
		json.Unmarshal(result, res)
		addr := res.Data.Detail.Addr
		plugin := res.Data.Plugin
		//utils.WarnOutput.Printf("[VULNERABLE] %s => %s\n", addr, plugin)
		logger.Warn("[VULNERABLE] %s => %s", addr, plugin)
	} else if rt.Type == "web_statistic" {
		webhook := &WebHookCountor{}
		json.Unmarshal(result, webhook)

		total := webhook.Data.NumFoundUrls
		scanned := webhook.Data.NumScannedUls
		pending := total - scanned
		ticker := time.NewTicker(10 * time.Second)

		//if no pending, stop output
		if pending != 0 {
			<-ticker.C
			utils.SuccessOut.Printf("Statistic: %d/%d, Pending: %d\n", scanned, total, pending)
		}

	} else {
		io.WriteString(w, "Error")
	}
}

func WebhookServer(wf *WebhookConfig) {

	http.HandleFunc(wf.Path, XrayWebhook)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", wf.Host, wf.Port), nil)
	if err != nil {
		log.Printf("%v", err)
	}
	fmt.Printf("[+]Start Webhook on %s:%d", wf.Host, wf.Port)
}

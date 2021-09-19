//从fofa搜索资产，将搜索结果返回给xray进行扫描
package fofa

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"snowball/config"
	"snowball/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	grumble "github.com/desertbit/grumble"
)

//fofa请求参数
type FofaRequest struct {
	Email   string
	Key     string
	Qbase64 string `json:"qbase64"`
	Page    int    `json:"page,omitempty"`
	Fields  string `json:"fields,omitempty"`
	Full    bool   `json:"full,omitempty"`
	Size    int    `json:"size,omitempty"`
}

//查询返回结果
type FofaResponse struct {
	Mode    string     `json:"mode"`
	Error   bool       `json:"error"`
	ErrMsg  string     `json:"errmsg"`
	Query   string     `json:"query"`
	Page    uint       `json:"page"`
	Size    uint       `json:"size"`
	Results [][]string `json:"results"`
}

const fofaURL = "https://fofa.so/api/v1/search/all?"

type FofaQuery struct {
	Context *grumble.Context
}

func (f *FofaQuery) Query() error {
	var email, key string
	fofa := &FofaRequest{}
	ctx := f.Context
	conf := &config.Fofa{}
	config.GetConfig(conf)
	email = ctx.Flags.String("email")
	key = ctx.Flags.String("key")

	if email != "" && key != "" {
		//命令行参数优先
		fofa.Email = email
		fofa.Key = key
	} else if conf.Email != "" && conf.Key != "" {
		fofa.Email = conf.Email
		fofa.Key = conf.Key
	} else {
		utils.WarnOutput.Println("[-]未配置fofa的email和key")
		return nil
	}

	if q := ctx.Flags.String("query"); q == "" {
		log.Println("查询内容不能为空")
		return nil
	} else {
		fofa.Qbase64 = q
	}

	fofa.Size = ctx.Flags.Int("size")

	res := fofa.RequestApi()
	if res == nil || res.Error {
		return nil
	} else {
		res.Results = RemoveRepeatElement(res.Results)
		var wg sync.WaitGroup
		ch := make(chan string, len(res.Results))
		for _, r := range res.Results {
			uri := utils.GenUrl(r[0])
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				ch <- GetWebTitle(u)
			}(uri)
			if ctx.Flags.Bool("scan") {
				//如果scan为false则不进行漏洞扫描，仅查询
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					AccessTargets(u)
				}(uri)
			}
		}
		go func() {
			wg.Wait()
			close(ch)
		}()

		for t := range ch {
			fmt.Printf("%s\n", t)
		}
		return nil
	}

}

func (req *FofaRequest) RequestApi() *FofaResponse {

	var qbase64 string
	resp := &FofaResponse{}

	if req.Qbase64 == "" {
		utils.WarnOutput.Println("[-]查询参数不能为空")
		return nil
	} else {
		qbase64 = base64.StdEncoding.EncodeToString([]byte(req.Qbase64))
	}

	//fofa查询请求参数
	params := url.Values{}
	params.Add("email", req.Email)
	params.Add("key", req.Key)
	params.Add("qbase64", qbase64)

	if req.Page != 0 {
		params.Add("page", strconv.Itoa(req.Page))
	}

	if req.Fields != "" {
		params.Add("fields", req.Fields)
	}

	if req.Full {
		params.Add("full", "true")
	}

	params.Add("size", strconv.Itoa(req.Size))

	request, err := http.NewRequest("GET", fofaURL, nil)

	request.URL.RawQuery = params.Encode()

	if err != nil {
		utils.WarnOutput.Println("[-]Request Error", request.RequestURI)
		return nil
	}

	response, err := utils.DoRequest(request)

	if err != nil || response.StatusCode != 200 {
		utils.WarnOutput.Println("[-]", err.Error())
		return nil
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		utils.WarnOutput.Println("[-]没有响应")
		return nil
	}

	json.Unmarshal(body, resp)

	return resp
}

func GetWebTitle(u string) string {
	var title string = ""
	doc, err := goquery.NewDocument(u)
	if err != nil {
		title = "Not Found"
		return fmt.Sprintf("%s => Title: %s", u, title)
	}
	title = doc.Find("title").Text()
	if strings.Trim(title, " ") == "" {
		title = "Not Found"
		return fmt.Sprintf("%s => Title: %s", u, title)
	}
	return fmt.Sprintf("%s => Title: %s", u, title)
}

//Access url ----(http proxy)------> xray ---(results)-----> snowball
func AccessTargets(u string) {
	conf := &config.Xray{}

	config.GetConfig(conf)

	host := conf.Host
	port := conf.Port

	if host == "" || port == 0 {
		log.Fatalln("未配置xray代理地址和端口")
	}

	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(fmt.Sprintf("http://%s:%d", host, port))
	}

	httpTransport := &http.Transport{
		Proxy:           proxy,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: httpTransport,
	}

	req, err := http.NewRequest("GET", u, nil)

	resp, err := client.Do(req)
	if err != nil {
		utils.WarnOutput.Printf("Access Error: %s\n", u)
		return
	}
	defer resp.Body.Close()

	ioutil.ReadAll(resp.Body)
}

func RemoveRepeatElement(resOld [][]string) [][]string {
	resNew := make([][]string, 0)
	for i := 0; i < len(resOld); i++ {
		flag := false
		for j := i + 1; j < len(resOld); j++ {
			if reflect.DeepEqual(resOld[i], resOld[j]) {
				flag = true
				break
			}
		}
		if !flag {
			resNew = append(resNew, resOld[i])
		}
	}
	return resNew
}

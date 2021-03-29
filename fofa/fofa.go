//从fofa搜索资产，将搜索结果返回给xray进行扫描
package fofa

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"snowball/utils"
	"snowball/xray"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	grumble "github.com/desertbit/grumble"
	yaml "gopkg.in/yaml.v2"
)

//定义fofa配置类型
//请求时需要在url中使用email和key
type FofaAuth struct {
	Email string `yaml:"email"`
	Key   string `yaml:"key"`
}

//fofa请求参数
type FofaRequestParams struct {
	FofaAuth `yaml:"fofa"`
	Qbase64  string `json:"qbase64"`
	Page     int    `json:"page"`
	Fields   string `json:"field"`
	Full     bool
	Size     uint
}

//查询返回结果
type FofaResults struct {
	Mode    string     `json:"mode"`
	Error   bool       `json:"error"`
	ErrMsg  string     `json:"errmsg"`
	Query   string     `json:"query"`
	Page    uint       `json:"page"`
	Size    uint       `json:"size"`
	Results [][]string `json:"results"`
}

func DoFofa(ctx *grumble.Context) error {
	req := &FofaRequestParams{}
	req.ReadConfigFile()

	//如果未指定fofa查询账号和密钥，则从配置文件中获取
	email := ctx.Flags.String("email")
	key := ctx.Flags.String("key")
	if key != "" && email != "" {
		req.Key = key
		req.Email = email
	}

	if q := ctx.Flags.String("query"); q == "" {
		log.Println("查询内容不能为空")
		return nil
	} else {
		req.Qbase64 = q
	}
	//fmt.Println(req.Qbase64)

	res := &FofaResults{}
	req.FofaRequest(res)
	if res.Error {
		log.Println("查询错误：", res.ErrMsg)
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

	}
	return nil
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
	xc := &xray.XrayConfig{}
	xc.ReadConfigFile()
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(fmt.Sprintf("http://%s:%d", xc.Host, xc.Port))
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
	// req, err = http.NewRequest("POST", u, nil)

	resp, err := client.Do(req)

	if err != nil {
		utils.WarnOutput.Printf("Access Error: %s\n", u)
		return
	}
	ioutil.ReadAll(resp.Body)
}

//读取配置文件中fofaAPI授权
func (conf *FofaRequestParams) ReadConfigFile() error {
	const ConfigFile = "config/config.yaml"
	buffer, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return err
	}
	yaml.Unmarshal(buffer, conf)
	return nil
}

func (req *FofaRequestParams) FofaRequest(res *FofaResults) error {
	//fofa查询API
	const fofaURL = "https://fofa.so/api/v1/search/all?"

	params := url.Values{}
	params.Add("email", req.Email)
	params.Add("key", req.Key)
	params.Add("qbase64", base64.StdEncoding.EncodeToString([]byte(req.Qbase64)))
	if req.Page != 0 {
		params.Add("page", strconv.Itoa(req.Page))
	}

	URL := fmt.Sprintf("%s%s", fofaURL, params.Encode())

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				DualStack: true,
			}).DialContext,
		},
		Timeout: 5 * time.Second,
	}
	response, err := client.Get(URL)
	if err != nil {
		utils.WarnOutput.Println("Create http request error")
		return err
	}
	if err != nil {
		log.Printf("Error:%v", err)
		return err
	}
	defer response.Body.Close()

	if err != nil {
		log.Printf("%v", err)
		return err
	}

	if response.StatusCode != 200 {
		log.Printf("返回响应码： %d", response.StatusCode)
		return nil
	}

	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, res)
	return nil
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

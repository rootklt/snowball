//交互式命令行

package interactor

import (
	//"fmt"

	"snowball/api"
	"snowball/fofa"
	"snowball/xray"

	"github.com/desertbit/grumble"
	"github.com/fatih/color"
)

var App = grumble.New(&grumble.Config{
	Name:                  "snowball",
	Description:           "利用fofa搜索信息，xray对目标进行漏洞探测",
	HistoryFile:           "/tmp/snowball.hist",
	Prompt:                "snowball>>",
	PromptColor:           color.New(color.FgGreen, color.Bold),
	HelpHeadlineColor:     color.New(color.FgGreen),
	HelpHeadlineUnderline: true,
	HelpSubCommands:       true,
	//全局命令参数
	Flags: func(f *grumble.Flags) {
		f.String("q", "query", "", "fofa搜索语法")
		f.Bool("v", "verbose", false, "enable verbose mode")
	},
})

func printLogo() {
	App.SetPrintASCIILogo(func(a *grumble.App) {
		a.Println("                              8             8 8 ")
		a.Println("                              8             8 8 ")
		a.Println(".oPYo. odYo. .oPYo. o   o   o 8oPYo. .oPYo. 8 8 ")
		a.Println("Yb..   8' `8 8    8 Y. .P. .P 8    8 .oooo8 8 8 ")
		a.Println("  'Yb. 8   8 8    8 `b.d'b.d' 8    8 8    8 8 8 ")
		a.Println("`YooP' 8   8 `YooP'  `Y' `Y'  `YooP' `YooP8 8 8 ")
		a.Println(":.....:..::..:.....:::..::..:::.....::.....:....")
		a.Println("::::::::::::::::::::::::::::::::::::::::::::::::")
		a.Println(":::::::::::::::::[by rootklt]:::::::::::::::::::")
		a.Println()
	})
}

func init() {
	printLogo()
}

func InterActor() {
	queryCmd := &grumble.Command{
		Name: "search",
		Help: "询查目标信息",
	}
	exitCmd := &grumble.Command{
		Name: "quit",
		Help: "quit process",
		Run: func(ctx *grumble.Context) error {
			xray.XrayStop()
			return nil
		},
	}
	xrayCmd := &grumble.Command{
		Name: "xray",
		Help: "xray漏洞扫描",
	}
	App.AddCommand(exitCmd)
	App.AddCommand(xrayCmd)
	App.AddCommand(queryCmd)
	XrayOption(xrayCmd)
	FofaCmd(queryCmd)
	grumble.Main(App)
}

func XrayOption(xrayCmd *grumble.Command) {
	xrayCmd.AddCommand(&grumble.Command{
		Name: "start",
		Help: "Start xray",
		Flags: func(f *grumble.Flags) {
			f.String("s", "host", "127.0.0.1", "fofa查询API密钥")
			f.Int("t", "port", 10810, "fofa查询API邮箱地址")
		},
		Run: func(ctx *grumble.Context) error {
			xray.XrayStart(ctx)
			return nil
		},
	})
	xrayCmd.AddCommand(&grumble.Command{
		Name: "stop",
		Help: "Stop xray",
		Run: func(ctx *grumble.Context) error {
			xray.XrayStop()
			return nil
		},
	})
	xrayCmd.AddCommand(&grumble.Command{
		Name: "scan",
		Help: "Xray 非代理扫描",
		Flags: func(f *grumble.Flags) {
			f.String("", "action", "webscan", "xray扫描功能选择，默认webscan")
			f.String("", "plugin", "phantasm", "扫描插件，默认为phantasm")
			f.String("", "poc", "", "指定pocs")
			f.String("", "url", "", "指定扫描目标url")
			f.String("", "output-fmt", "", "输出格式--html-output,--json--output, --webhook-output")
			f.String("", "file", "", "指定保存结果的文件")
		},
		Run: func(ctx *grumble.Context) error {
			xray.XrayScan(ctx)
			return nil
		},
	})
}

func FofaCmd(queryCmd *grumble.Command) {
	queryCmd.AddCommand(&grumble.Command{
		Name: "fofa",
		Help: "fofa平台询查",
		Flags: func(f *grumble.Flags) {
			f.String("k", "key", "", "fofa查询API密钥")
			f.String("e", "email", "", "fofa查询API邮箱地址")
			f.String("q", "query", "", "fofa查询语法")
			f.Uint("p", "page", 1, "查询页数")
			f.Bool("s", "scan", true, "默认不进行漏洞扫描")
			f.Int("z", "size", 100, "查询数量")
		},
		Run: func(ctx *grumble.Context) error {
			fofa := &fofa.FofaQuery{Context: ctx}
			api.DoQuery(fofa)
			return nil
		},
	})
}

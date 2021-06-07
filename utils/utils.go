package utils

import (
	"strings"

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

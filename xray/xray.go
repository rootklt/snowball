//启动、停止xray代理

package xray

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	//"os/signal"
	"runtime"
	"snowball/utils"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/desertbit/grumble"
	"gopkg.in/yaml.v2"
)

type Xray struct {
	Port    int      `yaml:"port"`
	Host    string   `yaml:"host"`
	Action  string   `yaml:"action"`
	Plugins []string `yaml:"plugins,flow"`
	Pocs    []string `yaml:"pocs,flow"`
}

type XrayConfig struct {
	Xray
}

const lockFile = "xray/xray.lock"
const ConfigFile = "config/config.yaml"

func (conf *XrayConfig) ReadConfigFile() {
	buffer, _ := ioutil.ReadFile(ConfigFile)
	yaml.Unmarshal(buffer, conf)
}

func XrayStart(ctx *grumble.Context) error {
	cf := &XrayConfig{}
	cf.ReadConfigFile()

	if host := ctx.Flags.String("host"); host != "" {
		cf.Host = host
	}
	if port := ctx.Flags.Int("port"); port != 0 {
		cf.Port = port
	}

	//webhook url
	wf := &WebhookConfig{}
	wf.ReadConfigFile()
	osType := runtime.GOOS
	xray := "xray/xray_darwin_amd64"
	if osType == "linux" {
		xray = "xray/xray_linux_amd64"
	} else if osType == "windows" {
		xray = "xray/xray_windows_amd64.exe"
	}
	if !isLockFileExist(lockFile) {
		var wg sync.WaitGroup
		c := make(chan *exec.Cmd)
		//启动xray监听服务
		wg.Add(1)
		go func(ch chan *exec.Cmd) {
			defer wg.Done()
			cmd := exec.Command(xray, "--config", "xray/config.yaml", cf.Action, "--plugins", strings.Join(cf.Plugins, ","), "--listen", fmt.Sprintf("%s:%d", cf.Host, cf.Port), "--webhook-output", fmt.Sprintf("http://%s:%d%s", wf.Host, wf.Port, wf.Path))
			cmd.Start()
			utils.SuccessOut.Printf("Xray start pid: %d\n", cmd.Process.Pid)
			lf, err := os.Create(lockFile)
			if err != nil {
				utils.WarnOutput.Println("Create Xray lock file Error")
			}
			defer lf.Close()
			lf.Write([]byte(strconv.Itoa(cmd.Process.Pid)))
			ch <- cmd
		}(c)

		go func(cmd chan *exec.Cmd) {
			cm := <-cmd
			wg.Wait()
			cm.Wait()
		}(c)
	}
	return nil
}

func isProcessExist() bool {
	cmd := exec.Command("ps", "c")
	output, _ := cmd.Output()
	fields := strings.Fields(string(output))
	for _, f := range fields {
		if strings.Contains(f, "xray_linux") {
			utils.SuccessOut.Printf("%s is Running\n", f)
			return true
		}
	}
	return false
}

func isLockFileExist(lockFile string) bool {
	var started bool = true
	pf, err := os.OpenFile(lockFile, os.O_RDWR, 0)
	if os.IsNotExist(err) {
		started = false
	} else {
		//如果lockfile存在，则要看pid是否是在运行
		pdf, err := ioutil.ReadAll(pf)
		pid, _ := strconv.Atoi(string(pdf))
		process, err := os.FindProcess(pid)

		if err != nil {
			utils.WarnOutput.Println("Not Found process in lock file, Start new process")
			process.Kill()
			os.Remove(lockFile)
			started = false
		} else {
			utils.SuccessOut.Printf("Xray is Running: [pid]%d\n", pid)
			started = true
		}

	}
	defer pf.Close()

	return started
}

func XrayStop() {
	pf, err := os.OpenFile(lockFile, os.O_RDWR, 0)
	if os.IsNotExist(err) {
		utils.WarnOutput.Println("lock file not exist")
	}
	defer pf.Close()
	pdf, err := ioutil.ReadAll(pf)
	if err != nil {
		utils.WarnOutput.Println("xray lock file read error")
		return
	}
	pid, err := strconv.Atoi(string(pdf))
	proccess, err := os.FindProcess(pid)
	if err != nil {
		utils.WarnOutput.Println("Not Found process")
		return
	}
	proccess.Signal(syscall.SIGTERM)
	utils.WarnOutput.Println("Xray Stoped...")
	defer os.Remove(lockFile)
	//go HandleSignal()
}

func XrayScan(ctx *grumble.Context) {
	url := ctx.Flags.String("url")
	if url == "" {
		utils.WarnOutput.Println("目标URL未指定")
		return
	}
	url = utils.GenUrl(url)

	action := ctx.Flags.String("action")
	poc := ctx.Flags.String("poc")
	if poc != "" {
		cmd := exec.Command("xray/xray_linux_amd64", action, "--poc", poc, "--url", url)
		cmd.Start()
	} else {
		plugin := ctx.Flags.String("plugin")
		cmd := exec.Command("xray/xray_linux_amd64", action, "--plugin", plugin, "--url", url)
		cmd.Start()
	}

}
/*
func HandleSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP)
	<-sigs
	XrayStop()
}
*/

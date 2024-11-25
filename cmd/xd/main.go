package main

import (
	"flag"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/donetkit/nps-client/client"
	"github.com/donetkit/nps-client/lib/common"
	"github.com/donetkit/nps-client/lib/install"
	"github.com/kardianos/service"
	"os"
	"runtime"
	"strings"
	"sync"
)

var sAddr, vKey, name string

var (
	logLevel   = flag.String("log_level", "2", "log level 0~7")
	logPath    = flag.String("log_path", "", "npc log path")
	serverInfo = flag.String("s", "", "server info")
)

func main() {
	//  -s 192.168.5.48:8124|123455|name
	flag.Parse()
	fmt.Println("start...")

	info := strings.SplitN(os.Getenv("s"), "_", -1)
	if len(info) != 3 {
		return
	}
	sAddr = info[0]
	vKey = info[1]
	name = info[2]

	RunNpc()
}

func RunNpc() {
	flag.Parse()
	logs.Reset()
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)

	if *logPath == "" {
		*logPath = common.GetNpcLogPath()
	}

	if common.IsWindows() {
		*logPath = strings.Replace(*logPath, "\\", "\\\\", -1)
	}

	logs.SetLogger(logs.AdapterFile, `{"level":`+*logLevel+`,"filename":"`+*logPath+`","daily":false,"maxlines":1000,"color":true}`)

	// init service
	options := make(service.KeyValue)
	svcConfig := &service.Config{
		Name:        "client",
		DisplayName: "客户端",
		Description: "一款轻量级、功能强大的内网穿透代理服务器",
		Option:      options,
	}
	if !common.IsWindows() {
		svcConfig.Dependencies = []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"}
		svcConfig.Option["SystemdScript"] = install.SystemdScript
		svcConfig.Option["SysvScript"] = install.SysvScript
	}
	for _, v := range os.Args[1:] {
		switch v {
		case "install", "start", "stop", "uninstall", "restart":
			continue
		}
		if !strings.Contains(v, "-service=") && !strings.Contains(v, "-debug=") {
			svcConfig.Arguments = append(svcConfig.Arguments, v)
		}
	}
	svcConfig.Arguments = append(svcConfig.Arguments, "-debug=false")
	prg := &npc{
		exit: make(chan struct{}),
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logs.Error(err, "service function disabled")
		run()
		// run without service
		wg := sync.WaitGroup{}
		wg.Add(1)
		wg.Wait()
		return
	}
	s.Run()
}

type npc struct {
	exit chan struct{}
}

func (p *npc) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *npc) Stop(s service.Service) error {
	close(p.exit)
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (p *npc) run() error {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logs.Warning("npc: panic serving %v: %v\n%s", err, string(buf))
		}
	}()
	run()
	select {
	case <-p.exit:
		logs.Warning("stop...")
	}
	return nil
}

func run() {
	var s = `[common]
server_addr=[server_addr]
conn_type=tcp
vkey=[vkey]
auto_reconnection=true
max_conn=500
flow_limit=100000
rate_limit=100000
crypt=true
compress=true
disconnect_timeout=60
remark=[name]`

	s = strings.ReplaceAll(s, "[server_addr]", sAddr)
	s = strings.ReplaceAll(s, "[name]", name)
	s = strings.ReplaceAll(s, "[vkey]", vKey)

	go client.StartFromString(s)

}

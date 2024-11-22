package client

import (
	"encoding/binary"
	"github.com/astaxie/beego/logs"
	"github.com/donetkit/nps-client/lib/common"
	"github.com/donetkit/nps-client/lib/config"
	"github.com/donetkit/nps-client/lib/version"
	"os"
	"path/filepath"
	"time"
)

func StartFromString(str string) {
	first := true
	cnf, err := config.NewConfigString(str)
	if err != nil || cnf.CommonConfig == nil {
		logs.Error("Config file %s loading error %s", "CONFIG", err.Error())
		os.Exit(0)
	}
	logs.Info("Loading configuration file %s successfully", "CONFIG")

	SetTlsEnable(cnf.CommonConfig.TlsEnable)
	logs.Info("the version of client is %s, the core version of client is %s,tls enable is %t", version.VERSION, version.GetVersion(), GetTlsEnable())
re:
	if first || cnf.CommonConfig.AutoReconnection {
		if !first {
			logs.Info("Reconnecting...")
			time.Sleep(time.Second * 5)
		}
	} else {
		return
	}
	first = false
	c, err := NewConn(cnf.CommonConfig.Tp, cnf.CommonConfig.VKey, cnf.CommonConfig.Server, common.WORK_CONFIG, cnf.CommonConfig.ProxyUrl)
	if err != nil {
		logs.Error(err)
		goto re
	}
	var isPub bool
	binary.Read(c, binary.LittleEndian, &isPub)

	// get tmp password
	var b []byte
	vkey := cnf.CommonConfig.VKey
	if isPub {
		// send global configuration to server and get status of config setting
		if _, err := c.SendInfo(cnf.CommonConfig.Client, common.NEW_CONF); err != nil {
			logs.Error(err)
			goto re
		}
		if !c.GetAddStatus() {
			logs.Error("the web_user may have been occupied!")
			goto re
		}

		if b, err = c.GetShortContent(16); err != nil {
			logs.Error(err)
			goto re
		}
		vkey = string(b)
	}
	os.WriteFile(filepath.Join(common.GetTmpPath(), "npc_vkey.txt"), []byte(vkey), 0600)

	//send hosts to server
	for _, v := range cnf.Hosts {
		if _, err := c.SendInfo(v, common.NEW_HOST); err != nil {
			logs.Error(err)
			goto re
		}
		if !c.GetAddStatus() {
			logs.Error(errAdd, v.Host)
			goto re
		}
	}

	//send  task to server
	for _, v := range cnf.Tasks {
		if _, err := c.SendInfo(v, common.NEW_TASK); err != nil {
			logs.Error(err)
			goto re
		}
		if !c.GetAddStatus() {
			logs.Error(errAdd, v.Ports, v.Remark)
			goto re
		}
		if v.Mode == "file" {
			//start local file server
			go startLocalFileServer(cnf.CommonConfig, v, vkey)
		}
	}

	//create local server secret or p2p
	for _, v := range cnf.LocalServer {
		go StartLocalServer(v, cnf.CommonConfig)
	}

	c.Close()
	if cnf.CommonConfig.Client.WebUserName == "" || cnf.CommonConfig.Client.WebPassword == "" {
		logs.Notice("web access login username:user password:%s", vkey)
	} else {
		logs.Notice("web access login username:%s password:%s", cnf.CommonConfig.Client.WebUserName, cnf.CommonConfig.Client.WebPassword)
	}
	NewRPClient(cnf.CommonConfig.Server, vkey, cnf.CommonConfig.Tp, cnf.CommonConfig.ProxyUrl, cnf, cnf.CommonConfig.DisconnectTime).Start()
	CloseLocalServer()
	goto re
}

package config

import (
	"github.com/donetkit/nps-client/lib/common"
	"strings"
)

func NewConfigString(str string) (c *Config, err error) {
	c = new(Config)
	if c.content, err = common.ParseStr(str); err != nil {
		return nil, err
	}
	if c.title, err = getAllTitle(c.content); err != nil {
		return
	}
	var nowIndex int
	var nextIndex int
	var nowContent string
	for i := 0; i < len(c.title); i++ {
		nowIndex = strings.Index(c.content, c.title[i]) + len(c.title[i])
		if i < len(c.title)-1 {
			nextIndex = strings.Index(c.content, c.title[i+1])
		} else {
			nextIndex = len(c.content)
		}
		nowContent = c.content[nowIndex:nextIndex]

		if strings.Index(getTitleContent(c.title[i]), "secret") == 0 && !strings.Contains(nowContent, "mode") {
			local := delLocalService(nowContent)
			local.Type = "secret"
			c.LocalServer = append(c.LocalServer, local)
			continue
		}
		//except mode
		if strings.Index(getTitleContent(c.title[i]), "p2p") == 0 && !strings.Contains(nowContent, "mode") {
			local := delLocalService(nowContent)
			local.Type = "p2p"
			c.LocalServer = append(c.LocalServer, local)
			continue
		}
		//health set
		if strings.Index(getTitleContent(c.title[i]), "health") == 0 {
			c.Healths = append(c.Healths, dealHealth(nowContent))
			continue
		}
		switch c.title[i] {
		case "[common]":
			c.CommonConfig = dealCommon(nowContent)
		default:
			if strings.Index(nowContent, "host") > -1 {
				h := dealHost(nowContent)
				h.Remark = getTitleContent(c.title[i])
				c.Hosts = append(c.Hosts, h)
			} else {
				t := dealTunnel(nowContent)
				t.Remark = getTitleContent(c.title[i])
				c.Tasks = append(c.Tasks, t)
			}
		}
	}
	return
}

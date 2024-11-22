//go:build windows
// +build windows

package proxy

import (
	"github.com/donetkit/nps-client/lib/conn"
)

func HandleTrans(c *conn.Conn, s *TunnelModeServer) error {
	return nil
}

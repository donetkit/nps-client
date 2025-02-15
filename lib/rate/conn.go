package rate

import (
	"github.com/donetkit/nps-client/lib/file"
	"io"
)

type rateConn struct {
	conn io.ReadWriteCloser
	rate *file.Rate
}

func NewRateConn(conn io.ReadWriteCloser, rate *file.Rate) io.ReadWriteCloser {
	return &rateConn{
		conn: conn,
		rate: rate,
	}
}

func (s *rateConn) Read(b []byte) (n int, err error) {
	n, err = s.conn.Read(b)
	if s.rate != nil {
		s.rate.Get(int64(n))
	}
	return
}

func (s *rateConn) Write(b []byte) (n int, err error) {
	n, err = s.conn.Write(b)
	if s.rate != nil {
		s.rate.Get(int64(n))
	}
	return
}

func (s *rateConn) Close() error {
	return s.conn.Close()
}

package file

import (
	"sync/atomic"
	"time"
)

type Rate struct {
	bucketSize        int64
	bucketSurplusSize int64
	bucketAddSize     int64
	stopChan          chan bool
	NowRate           int64
	c                 *Client
}

func NewRate(addSize int64, client *Client) *Rate {
	return &Rate{
		bucketSize:        addSize * 2,
		bucketSurplusSize: 0,
		bucketAddSize:     addSize,
		stopChan:          make(chan bool),
		c:                 client,
	}
}

func (s *Rate) Start() {
	go s.session()
}

func (s *Rate) add(size int64) {
	if res := s.bucketSize - s.bucketSurplusSize; res < s.bucketAddSize {
		atomic.AddInt64(&s.bucketSurplusSize, res)
		return
	}
	atomic.AddInt64(&s.bucketSurplusSize, size)
}

// 回桶
func (s *Rate) ReturnBucket(size int64) {
	s.add(size)
}

// 停止
func (s *Rate) Stop() {
	s.stopChan <- true
}

func (s *Rate) Get(size int64) {
	if s.bucketSurplusSize >= size {
		atomic.AddInt64(&s.bucketSurplusSize, -size)
		return
	}
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-ticker.C:
			if s.bucketSurplusSize >= size {
				atomic.AddInt64(&s.bucketSurplusSize, -size)
				ticker.Stop()
				return
			}
		}
	}
}

func (s *Rate) session() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, err := GetDb().GetClient(s.c.Id)
			if err != nil {
				//fmt.Println("AAAAA---------(s *Rate) session()-------err= ", err, s.c.Id)
				return
			}
			if rs := s.bucketAddSize - s.bucketSurplusSize; rs > 0 {
				s.NowRate = rs
			} else {
				s.NowRate = s.bucketSize - s.bucketSurplusSize
			}
			s.add(s.bucketAddSize)
		case <-s.stopChan:
			return
		}
	}
}

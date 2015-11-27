package source

import (
	"container/list"
	"errors"
	"io"
	"sync/atomic"

	"github.com/golang/glog"
)

const (
	DefaultBufCacheLength = 4096 * 8
	DefautLength          = DefaultCacheMaxLength
	DefaultSize           = 1024 * 1024 * 20
)

type Consumer struct {
	bufChan chan *msg
	length  int64
	size    int64

	maxLen  int64
	maxSize int64

	headBuf []byte

	startTime uint64
	preLength uint32
	sourcer   *Sourcer

	e *list.Element

	isClosed bool
}

var notSource = errors.New("source not found.")

func NewConsumer(key string) (*Consumer, error) {
	glog.Info("[Hang]: get source hang!")
	s, ok := Sources.Get(key)
	glog.Info("[Hang]: get source hang done!")
	if !ok {
		return nil, notSource
	}
	consumer := &Consumer{
		bufChan: make(chan *msg, DefaultBufCacheLength),
		sourcer: s,
		headBuf: make([]byte, 15),
		maxLen:  DefautLength,
		maxSize: DefaultSize,
	}
	// add consumer to sourcer.
	glog.Info("[Hang]: get consumer hang!")
	e, err := consumer.sourcer.addConsumer(consumer)
	if err != nil {
		return nil, err
	}
	glog.Info("[Hang]: get consumer hang done!")
	consumer.e = e
	return consumer, nil
}

func (c *Consumer) Close() {
	c.sourcer.Lock()
	c.sourcer.delConsumer(c.e)
	c.sourcer.Unlock()
}

// Range consumer sourcer must call read lock
func (c *Consumer) addMsg(m *msg) {
	length := atomic.AddInt64(&c.length, 0)
	size := atomic.AddInt64(&c.size, 0)
	if length == c.maxLen || size > c.maxSize {
		return
	}
	atomic.AddInt64(&c.length, 1)
	atomic.AddInt64(&c.size, int64(m.Header.PayloadLength))
	c.bufChan <- m
}

func (c *Consumer) Live(w io.Writer) error {
	var (
		err error
		// get first msg.
		isFirst = true
	)
	if err = c.writeFlvHead(w); err != nil {
		return err
	}

	for {
		m, ok := <-c.bufChan
		if !ok {
			return nil
		}

		// glog.Info(c.length, c.size)
		atomic.AddInt64(&c.length, -1)
		atomic.AddInt64(&c.size, -int64(m.Header.PayloadLength))
		// write to w.
		ok = m.getFlvTagHead(c.startTime, c.preLength, c.headBuf)
		if !ok {
			continue
		}
		if isFirst {
			c.startTime = m.Header.Timestamp
			c.preLength = m.Header.PayloadLength
			isFirst = false
		}
		c.preLength = m.Header.PayloadLength
		if _, err = w.Write(c.headBuf); err != nil {
			return err
		}
		if _, err = w.Write(m.Payload); err != nil {
			return err
		}
	}
}

func (c *Consumer) writeFlvHead(w io.Writer) error {
	glog.Info("<<<<<<<write flv file head........>>>>>>>>>")
	var (
		err error
		ok  bool
	)
	if _, err = w.Write(c.sourcer.flvHead); err != nil {
		return err
	}
	ok = c.sourcer.metaHead.getFlvTagHead(0, 0, c.headBuf)
	if !ok {
		return notGetMeta
	}
	if _, err = w.Write(c.headBuf); err != nil {
		return err
	}
	if _, err = w.Write(c.sourcer.metaHead.Payload); err != nil {
		return err
	}
	if c.sourcer.audioMeta != nil {
		glog.Info("trace: add audio meta")
		ok = c.sourcer.audioMeta.getFlvTagHead(0, 0, c.headBuf)
		if !ok {
			return notGetMeta
		}
		if _, err = w.Write(c.headBuf); err != nil {
			return err
		}
		if _, err = w.Write(c.sourcer.audioMeta.Payload); err != nil {
			return err
		}
	}
	if c.sourcer.videoMeta != nil {
		glog.Info("trace: add video meta")
		ok = c.sourcer.videoMeta.getFlvTagHead(0, 0, c.headBuf)
		if !ok {
			return notGetMeta
		}
		if _, err = w.Write(c.headBuf); err != nil {
			return err
		}
		if _, err = w.Write(c.sourcer.videoMeta.Payload); err != nil {
			return err
		}
	}
	return nil
}

package source

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/Alienero/IamServer/rtmp"

	"github.com/elobuff/goamf"
)

const DefaultCacheMaxLength = 1024 * 1024 * 500

var (
	flvHeadAudio = []byte{'F', 'L', 'V', 0x01,
		0x04,
		0x00, 0x00, 0x00, 0x09}

	flvHeadVideo = []byte{'F', 'L', 'V', 0x01,
		0x01,
		0x00, 0x00, 0x00, 0x09}

	flvHeadBoth = []byte{'F', 'L', 'V', 0x01,
		0x05,
		0x00, 0x00, 0x00, 0x09}
)

type Sourcer struct {
	msgs      *list.List
	flvHead   []byte
	metaHead  *msg
	preTagLen uint32
	// for gc
	sync.RWMutex
	cond *sync.Cond
	// number of goroutinue on hang stat.
	HangWait uint32
	// total cached time.
	cachedLength uint64
	// head time index.
	// headTime uint64

	isClosed bool
}

// New->Flv head -> Meta head-> transport -> Close.
func NewSourcer() *Sourcer {
	return &Sourcer{
		msgs: list.New(),
		cond: sync.NewCond(new(sync.Mutex)),
	}
}

// set meta and headTime.
func (s *Sourcer) SetMeta(message *rtmp.Message) error {
	decoder := amf.NewDecoder()
	reader := bytes.NewReader(message.Payload)
	l := reader.Len()
	for {
		v, err := decoder.DecodeAmf0(reader)
		if err != nil {
			return err
		}
		if str := v.(string); str != "@setDataFrame" {
			meta := message.Payload[int(message.Header.PayloadLength)-l:]
			println("trace:get meta head.")
			s.metaHead = newMsg(message)
			s.metaHead.Payload = meta
			s.metaHead.Header.PayloadLength = uint32(len(meta))
			return nil
		}
		l = reader.Len()
	}
}

func (s *Sourcer) SetFlvHead() {
	// default.
	s.flvHead = flvHeadBoth
}

func (s *Sourcer) Close() error {
	s.Lock()
	s.isClosed = true
	s.Unlock()
	s.cond.Broadcast()
	return nil
}

func (s *Sourcer) HandleMsg(message *rtmp.Message) {
	m := new(msg)
	m.Message = *message
	m.preLength = s.preTagLen
	s.preTagLen = message.Header.PayloadLength

	// add message to the end of msgs list.
	s.msgs.PushBack(m)
	s.RLock()
	n := atomic.AddUint32(&s.HangWait, 0)
	s.Unlock()
	if n > 0 {
		atomic.AddUint32(&s.HangWait, -n)
		// some was hang.
		s.cond.Broadcast()
	}

	// check gc.
	s.cachedLength += uint64(m.Header.PayloadLength)
	if s.cachedLength > DefaultCacheMaxLength {
		s.Lock()
		for p := s.msgs.Front(); p != nil || s.cachedLength < (DefaultCacheMaxLength*0.7); p = p.Next() {
			s.cachedLength -= uint64(p.Value.(*msg).Header.PayloadLength)
		}
		s.msgs.Front().Value.(*msg).preLength = 0
		s.Unlock()
	}

}

// If close,unless not return.
// TODO: ctrl the transport speed.
func (s *Sourcer) Live(w io.Writer) error {
	if _, err := w.Write(s.flvHead); err != nil {
		return err
	}
	metaHead, ok := s.metaHead.getFlvTagHead(0)
	if !ok {
		return errors.New("can't get flv meta data.")
	}
	if _, err := w.Write(metaHead); err != nil {
		return err
	}
	if _, err := w.Write(s.metaHead.Payload); err != nil {
		return err
	}
	var (
		startTime uint64
		node      *list.Element
		isFirst   = true
	)

	for {
		// check isClosed.
		s.RLock()
		if s.isClosed {
			s.RLock()
			println("is close, read return.")
			return nil
		}
		if node == nil {
			// get new node.
			n := s.msgs.Len()
			if n == 0 {
				atomic.AddUint32(&s.HangWait, 1)
				s.Unlock()
				s.cond.L.Lock()
				s.cond.Wait()
				s.cond.L.Unlock()
				continue
			}
			n = int(float64(n) * 0.6)
			if n == 0 {
				n = 1
			}
			node = s.msgs.Front()
			if n == 1 {
				s.RUnlock()
				continue
			}
			for i := 1; i < n; i++ {
				node = node.Next()
			}
			s.RUnlock()
		} else {
			// get startTime && node
			if isFirst {
				startTime = node.Value.(*msg).Header.Timestamp
				isFirst = false
			}
			s.RUnlock()
		}
		m := node.Value.(*msg)
		tagHead, ok := m.getFlvTagHead(startTime)
		if ok {
			if _, err := w.Write(tagHead); err != nil {
				return err
			}
			if _, err := w.Write(m.Payload); err != nil {
				return err
			}
		}
		s.RLock()
		node = node.Next()
		s.RUnlock()
		if node == nil {
			atomic.AddUint32(&s.HangWait, 1)
			s.cond.L.Lock()
			s.cond.Wait()
			s.cond.L.Unlock()
		}
	}
}

type msg struct {
	rtmp.Message
	preLength uint32
	refer     int
}

func newMsg(m *rtmp.Message) *msg {
	ms := new(msg)
	ms.Message = *m
	return ms
}

func (m *msg) getFlvTagHead(startTime uint64) ([]byte, bool) {
	b := make([]byte, 0, 15)
	b = append(b, uint32ToBytes(m.preLength)...)
	b = append(b, m.Header.MessageType)
	b = append(b, uint32ToBytes(m.Header.PayloadLength)[1:]...)

	indexTime := m.Header.Timestamp - startTime
	if indexTime < 0 {
		return nil, false
	}
	t := uint32ToBytes(uint32(indexTime))

	b = append(b, t[1:]...)
	b = append(b, t[0])

	return append(b, []byte{0, 0, 0}...), true
}

func uint32ToBytes(l uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(l))
	return b
}

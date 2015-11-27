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
	"github.com/golang/glog"
)

const DefaultCacheMaxLength = 1024 * 1024 * 20

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

type sourceManage struct {
	dict map[string]*Sourcer
	sync.RWMutex
}

var (
	sourExist = errors.New("source exits.")
	Sources   = &sourceManage{
		dict: make(map[string]*Sourcer),
	}
)

func (sm *sourceManage) Set(key string) (*Sourcer, error) {
	sm.Lock()
	defer sm.Unlock()
	_, ok := sm.dict[key]
	if ok {
		return nil, sourExist
	}
	s := NewSourcer(key)
	sm.dict[key] = s
	return s, nil
}

func (sm *sourceManage) Get(key string) (*Sourcer, bool) {
	sm.RLock()
	defer sm.RUnlock()
	s, ok := sm.dict[key]
	return s, ok
}

func (sm *sourceManage) Delete(key string) {
	sm.Lock()
	delete(sm.dict, key)
	sm.Unlock()
}

type Sourcer struct {
	msgs      *list.List
	flvHead   []byte
	metaHead  *msg
	audioMeta *msg
	videoMeta *msg
	// preTagLen uint32
	// for gc
	sync.RWMutex
	// number of goroutinue on hang stat.
	HangWait int32
	// total cached time.
	cachedLength uint64
	key          string

	isClosed  bool
	isRunning bool

	seq        uint64
	signalChan chan struct{}
}

// New->Flv head -> Meta head-> transport -> Close.
func NewSourcer(key string) *Sourcer {
	return &Sourcer{
		msgs:       list.New(),
		key:        key,
		signalChan: make(chan struct{}, 5000),
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
			glog.Info("trace:get meta head.")
			s.metaHead = newMsg(message)
			s.metaHead.Payload = meta
			s.metaHead.Header.PayloadLength = uint32(len(meta))
			return nil
		}
		l = reader.Len()
	}
}

func (s *Sourcer) SetAudioMeta(m *rtmp.Message) {
	s.audioMeta = newMsg(m)
}

func (s *Sourcer) SetVideoMeta(m *rtmp.Message) {
	s.videoMeta = newMsg(m)
}

func (s *Sourcer) SetFlvHead() {
	// default.
	s.flvHead = flvHeadBoth
}

func (s *Sourcer) Run() {
	s.Lock()
	s.isRunning = true
	s.Unlock()
}

func (s *Sourcer) Close() error {
	s.Lock()
	s.isClosed = true
	s.Unlock()
	s.wakeAll()
	return nil
}

// only, server not closed.
// before you should call s.RLock().
func (s *Sourcer) wait() {
	isClosed := s.isClosed
	if !isClosed {
		atomic.AddInt32(&s.HangWait, 1)
		s.RUnlock()
		<-s.signalChan
	} else {
		s.RUnlock()
	}
}

func (s *Sourcer) wakeAll() {
	n := atomic.AddInt32(&s.HangWait, 0)
	if n > 0 {
		s.wakeImp(n)
		atomic.AddInt32(&s.HangWait, -n)
	}
}

func (s *Sourcer) wakeImp(n int32) {
	for i := 0; i < int(n); i++ {
		s.signalChan <- struct{}{}
	}
}

func (s *Sourcer) HandleMsg(message *rtmp.Message) {
	m := new(msg)
	m.Message = *message
	s.seq++ // not need thread safe.
	m.seq = s.seq

	// add message to the end of msgs list.
	s.msgs.PushBack(m)

	// wake up all waitter.
	s.wakeAll()

	// check gc.
	s.cachedLength += uint64(m.Header.PayloadLength)
	if s.cachedLength > DefaultCacheMaxLength {
		s.Lock()
		p := s.msgs.Front()
		i := DefaultCacheMaxLength * 0.7
		for p != nil && s.cachedLength > uint64(i) {
			temp := p
			p = p.Next()
			s.msgs.Remove(temp)
			s.cachedLength -= uint64(temp.Value.(*msg).Header.PayloadLength)
		}
		s.Unlock()
	}
}

var notGetMeta = errors.New("can't get flv meta data")
var notRun = errors.New("source not running")

// If close,unless not return.
// TODO: add time meta.
// TODO: ctrl the transport speed.
func (s *Sourcer) Live(w io.Writer) error {
	s.RLock()
	if !s.isRunning {
		s.RUnlock()
		return notRun
	}
	s.RUnlock()

	var (
		err    error
		hasSeq uint64
		ok     bool

		// buf total tag's head.
		buf = make([]byte, 15)
	)

	if _, err = w.Write(s.flvHead); err != nil {
		return err
	}
	ok = s.metaHead.getFlvTagHead(0, 0, buf)
	if !ok {
		return notGetMeta
	}
	if _, err = w.Write(buf); err != nil {
		return err
	}
	if _, err = w.Write(s.metaHead.Payload); err != nil {
		return err
	}
	if s.audioMeta != nil {
		glog.Info("trace: add audio meta")
		ok = s.audioMeta.getFlvTagHead(0, 0, buf)
		if !ok {
			return notGetMeta
		}
		if _, err = w.Write(buf); err != nil {
			return err
		}
		if _, err = w.Write(s.audioMeta.Payload); err != nil {
			return err
		}
	}
	if s.videoMeta != nil {
		glog.Info("trace: add video meta")
		ok = s.videoMeta.getFlvTagHead(0, 0, buf)
		if !ok {
			return notGetMeta
		}
		if _, err = w.Write(buf); err != nil {
			return err
		}
		if _, err = w.Write(s.videoMeta.Payload); err != nil {
			return err
		}
	}
	var (
		startTime uint64
		node      *list.Element
		isFirst          = true
		preLength uint32 = s.metaHead.Header.PayloadLength
	)

	for {
		// check isClosed.
		s.RLock()
		if s.isClosed {
			s.RUnlock()
			return nil
		}
		// there lock for safe to get node, without nil node.
		if node == nil {
			// get new node.
			if s.msgs.Len() == 0 || s.seq == hasSeq {
				s.wait() // it will call s.RUnlock().
				continue
			}
			node = s.msgs.Back()
			// get startTime && node
			if isFirst {
				startTime = node.Value.(*msg).Header.Timestamp
				isFirst = false
			}
		}
		s.RUnlock()

		m := node.Value.(*msg)
		hasSeq = m.seq
		ok := m.getFlvTagHead(startTime, preLength, buf)
		if ok {
			preLength = m.Header.PayloadLength
			if _, err = w.Write(buf); err != nil {
				return err
			}
			if _, err = w.Write(m.Payload); err != nil {
				return err
			}
		}
		// safe get node.
		s.RLock()
		node = node.Next()
		if node == nil {
			s.wait() // it will call s.RUnlock().
		} else {
			s.RUnlock()
		}
	}
}

// TODO: finish consumer. Imp it.
type consumer struct {
	src        *Sourcer
	tagHeadBuf []byte
	startTime  uint64
}

type msg struct {
	rtmp.Message
	refer int
	seq   uint64
}

func newMsg(m *rtmp.Message) *msg {
	ms := new(msg)
	ms.Message = *m
	return ms
}

func (m *msg) getFlvTagHead(startTime uint64, preLength uint32, buf []byte) bool {
	uint32ToBytes(preLength, buf[0:4])
	buf[4] = m.Header.MessageType
	temp := buf[5]
	uint32ToBytes(m.Header.PayloadLength, buf[5:9])
	buf[5] = temp

	indexTime := m.Header.Timestamp - startTime
	if indexTime < 0 {
		return false
	}
	uint32ToBytes(uint32(indexTime), buf[9:13])
	high := buf[9]
	buf[9] = buf[10]
	buf[10] = buf[11]
	buf[11] = buf[12]
	buf[12] = high

	return true
}

func uint32ToBytes(l uint32, buf []byte) {
	binary.BigEndian.PutUint32(buf, uint32(l))
}

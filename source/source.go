package source

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"sync"

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

	isClosed  bool
	isRunning bool

	seq uint64

	consumers *list.List
}

// New->Flv head -> Meta head-> transport -> Close.
func NewSourcer(key string) *Sourcer {
	return &Sourcer{
		msgs:      list.New(),
		consumers: list.New(),
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
	s.delAllConsumer()
	s.Unlock()
	return nil
}

func (s *Sourcer) addConsumer(c *Consumer) (*list.Element, error) {
	s.Lock()
	defer s.Unlock()
	if s.isClosed {
		return nil, errors.New("source is closed.")
	}
	return s.consumers.PushBack(c), nil
}

// Consumer call it.
func (s *Sourcer) delConsumer(e *list.Element) {
	s.Lock()
	defer s.Unlock()
	c := s.consumers.Remove(e).(*Consumer)
	close(c.bufChan)
}

// Only Sourcer's close method call it.
func (s *Sourcer) delAllConsumer() {
	for node := s.consumers.Front(); node != nil; node = node.Next() {
		close(node.Value.(*Consumer).bufChan)
	}
}

func (s *Sourcer) HandleMsg(message *rtmp.Message) {
	m := new(msg)
	m.Message = *message
	s.seq++ // not need thread safe.
	s.RLock()
	for node := s.consumers.Front(); node != nil; node = node.Next() {
		node.Value.(*Consumer).addMsg(m)
	}
	s.RUnlock()
}

var notGetMeta = errors.New("can't get flv meta data")
var notRun = errors.New("source not running")

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
	temp := m.Header.MessageType
	uint32ToBytes(m.Header.PayloadLength, buf[4:8])
	buf[4] = temp

	indexTime := m.Header.Timestamp - startTime
	if indexTime < 0 {
		return false
	}
	uint32ToBytes(uint32(indexTime), buf[8:12])
	high := buf[8]
	buf[8] = buf[9]
	buf[9] = buf[10]
	buf[10] = buf[11]
	buf[11] = high

	return true
}

func uint32ToBytes(l uint32, buf []byte) {
	binary.BigEndian.PutUint32(buf, uint32(l))
}

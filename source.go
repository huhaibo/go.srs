package main

import (
	"container/list"
	"sync"

	"github.com/Alienero/IamServer/rtmp"
)

var source_pool map[string]*SrsSource = map[string]*SrsSource{}

/**
* live streaming source.
 */
type SrsSource struct {
	// the identified request from client.
	req *rtmp.Request
	// the consumer list
	consumers      *list.List
	consumers_lock *sync.Mutex
	/**
	* the sample rate of audio in metadata.
	 */
	sample_rate int
	/**
	* the video frame rate in metadata.
	 */
	frame_rate int
}

/**
* find stream by vhost/app/stream.
* @param req the client request.
* @return the matched source, never be NULL.
* @remark stream_url should without port and schema.
 */
func FindSrsSource(req *rtmp.Request) *SrsSource {
	stream_url := "/" + req.App + "/" + req.Stream

	if _, ok := source_pool[stream_url]; !ok {
		r := &SrsSource{}
		r.req = req
		r.consumers = list.New()
		r.consumers_lock = &sync.Mutex{}

		source_pool[stream_url] = r
	}
	return source_pool[stream_url]
}
func (r *SrsSource) CreateConsumer() *SrsConsumer {
	r.consumers_lock.Lock()
	defer r.consumers_lock.Unlock()

	v := NewSrsConsumer(r)
	v.elem = r.consumers.PushBack(v)
	return v
}
func (r *SrsSource) RemoveConsumer(v *SrsConsumer) {
	r.consumers_lock.Lock()
	defer r.consumers_lock.Unlock()

	if v.elem != nil {
		r.consumers.Remove(v.elem)
	}
}
func (r *SrsSource) OnAudio(msg *rtmp.Message) (err error) {
	r.consumers_lock.Lock()
	defer r.consumers_lock.Unlock()

	// SRS_HLS
	// TODO: FIXME: implements it.

	// copy to all consumer
	for p := r.consumers.Front(); p != nil; p = p.Next() {
		p := p.Value.(*SrsConsumer)
		if err = p.OnMessage(msg.Copy(), r.sample_rate, r.frame_rate); err != nil {
			return
		}
	}
	return
}
func (r *SrsSource) OnVideo(msg *rtmp.Message) (err error) {
	r.consumers_lock.Lock()
	defer r.consumers_lock.Unlock()

	// SRS_HLS
	// TODO: FIXME: implements it.

	// copy to all consumer
	for p := r.consumers.Front(); p != nil; p = p.Next() {
		p := p.Value.(*SrsConsumer)
		if err = p.OnMessage(msg.Copy(), r.sample_rate, r.frame_rate); err != nil {
			return
		}
	}
	return
}

/**
* the consumer for SrsSource, that is a play client.
 */
type SrsConsumer struct {
	source *SrsSource
	msgs   chan *rtmp.Message
	elem   *list.Element
}

func NewSrsConsumer(source *SrsSource) *SrsConsumer {
	r := &SrsConsumer{}
	r.source = source
	// TODO: FIXME: use buffered channel
	r.msgs = make(chan *rtmp.Message, 1000)
	return r
}
func (r *SrsConsumer) Messages() chan *rtmp.Message {
	return r.msgs
}
func (r *SrsConsumer) OnMessage(msg *rtmp.Message, tba int, tbv int) (err error) {
	r.msgs <- msg
	return
}

/**
* close the consumer, for example, client play another source.
 */
func (r *SrsConsumer) Close() (err error) {
	r.source.RemoveConsumer(r)
	r.source = nil
	close(r.msgs)
	return
}

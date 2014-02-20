// The MIT License (MIT)
//
// Copyright (c) 2014 winlin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"github.com/winlinvip/go.rtmp/rtmp"
	"container/list"
	"fmt"
)

var source_pool map[string]*SrsSource = map[string]*SrsSource{}

/**
* live streaming source.
*/
type SrsSource struct {
	// the identified request from client.
	req *rtmp.Request
	// the consumer list
	consumers *list.List
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
func FindSrsSource(req *rtmp.Request) (*SrsSource) {
	stream_url := req.StreamUrl()
	fmt.Println("discovery source", stream_url)
	if _, ok := source_pool[stream_url]; !ok {
		r := &SrsSource{}
		r.req = req
		r.consumers = list.New()

		source_pool[stream_url] = r
	}
	return source_pool[stream_url]
}
func (r *SrsSource) CreateConsumer() (*SrsConsumer) {
	v := NewSrsConsumer(r)
	v.elem = r.consumers.PushBack(v)
	return v
}
func (r *SrsSource) RemoveConsumer(v *SrsConsumer){
	if v.elem != nil {
		r.consumers.Remove(v.elem)
	}
}
func (r *SrsSource) OnAudio(msg *rtmp.Message) (err error) {
	// SRS_HLS
	// TODO: FIXME: implements it.

	// copy to all consumer
	for p := r.consumers.Front(); p != nil; p = p.Next() {
		p := p.Value.(*SrsConsumer)
		if err = p.OnMessage(msg, r.sample_rate, r.frame_rate); err != nil {
			return
		}
	}
	return
}
func (r *SrsSource) OnVideo(msg *rtmp.Message) (err error) {
	// SRS_HLS
	// TODO: FIXME: implements it.

	// copy to all consumer
	for p := r.consumers.Front(); p != nil; p = p.Next() {
		p := p.Value.(*SrsConsumer)
		if err = p.OnMessage(msg, r.sample_rate, r.frame_rate); err != nil {
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
	msgs chan *rtmp.Message
	elem *list.Element
}
func NewSrsConsumer(source *SrsSource) (*SrsConsumer) {
	r := &SrsConsumer{}
	r.source = source
	// TODO: FIXME: use buffered channel
	r.msgs = make(chan *rtmp.Message, 1000)
	return r
}
func (r *SrsConsumer) OnMessage(msg *rtmp.Message, tba int, tbv int) (err error) {
	// TODO: FIXME: drop if overflow.
	r.msgs <- msg
	return
}
/**
* close the consumer, for example, client play another source.
 */
func (r *SrsConsumer) Close() (err error) {
	r.source.RemoveConsumer(r)
	close(r.msgs)
	r.source = nil
	return
}

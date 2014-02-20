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
	"net"
	"fmt"
	"github.com/winlinvip/go.rtmp/rtmp"
	"time"
)

// default stream id for response the createStream request.
const SRS_DEFAULT_SID = 1

/**
* the response info for srs.
 */
type SrsResponse struct {
	stream_id int
}
func NewSrsResponse() (*SrsResponse) {
	r := &SrsResponse{}
	r.stream_id = SRS_DEFAULT_SID
	return r
}
// interface RtmpStreamIdGenerator
func (r *SrsResponse) StreamId() (n int) {
	return r.stream_id
}

/**
* the client provides the main logic control for RTMP clients.
*/
type SrsClient struct {
	conn *net.TCPConn
	rtmp rtmp.Server
	req *rtmp.Request
	res *SrsResponse
	consumer *SrsConsumer
}
func NewSrsClient(conn *net.TCPConn) (r *SrsClient, err error) {
	r = &SrsClient{}
	r.conn = conn
	r.res = NewSrsResponse()

	if r.rtmp, err = rtmp.NewServer(conn); err != nil {
		return
	}
	r.req = rtmp.NewRequest()
	return
}

func (r *SrsClient) do_cycle() (err error) {
	if err = r.rtmp.Handshake(); err != nil {
		return
	}

	if err = r.rtmp.ConnectApp(r.req); err != nil {
		return
	}
	fmt.Printf("request, tcUrl=%v(vhost=%v, app=%v), AMF%v, pageUrl=%v, swfUrl=%v\n",
		r.req.TcUrl, r.req.Vhost, r.req.App, r.req.ObjectEncoding, r.req.PageUrl, r.req.SwfUrl)

	// check_vhost
	// TODO: FIXME: implements it

	err = r.service_cycle()
	// on_close
	return
}
func (r *SrsClient) service_cycle() (err error) {
	ack_size := uint32(2.5 * 1000 * 1000)
	if err = r.rtmp.SetWindowAckSize(ack_size); err != nil {
		return
	}
	fmt.Printf("set window ack size to %v\n", ack_size)

	bandwidth, bw_type := uint32(2.5 * 1000 * 1000), byte(2)
	if err = r.rtmp.SetPeerBandwidth(bandwidth, bw_type); err != nil {
		return
	}
	fmt.Printf("set bandwidth to %v, type=%v\n", bandwidth, bw_type)

	// do bandwidth test if connect to the vhost which is for bandwidth check.
	// TODO: FIXME: implements it

	extra_data := []map[string]string {
		{ "srs_sig": RTMP_SIG_SRS_KEY },
		{ "srs_server": RTMP_SIG_SRS_KEY + " " + RTMP_SIG_SRS_VERSION + " (" + RTMP_SIG_SRS_URL_SHORT + ")" },
		{ "srs_license": RTMP_SIG_SRS_LICENSE },
		{ "srs_role": RTMP_SIG_SRS_ROLE },
		{ "srs_url": RTMP_SIG_SRS_URL },
		{ "srs_version": RTMP_SIG_SRS_VERSION },
		{ "srs_site": RTMP_SIG_SRS_WEB },
		{ "srs_email": RTMP_SIG_SRS_EMAIL },
		{ "srs_copyright": RTMP_SIG_SRS_COPYRIGHT },
		{ "srs_primary_authors": RTMP_SIG_SRS_PRIMARY_AUTHROS },
	}
	if err = r.rtmp.ReponseConnectApp(r.req, "", extra_data); err != nil {
		return
	}
	fmt.Printf("response connect app success\n")

	if err = r.rtmp.CallOnBWDone(); err != nil {
		return
	}
	fmt.Printf("call client as onBWDone()\n")

	for {
		err = r.stream_service_cycle()

		// stream service must terminated with error, never success.
		if err == nil {
			fmt.Println("stream service complete success, re-identify it")
			continue
		}

		// when not system control error, fatal error, return.
		if !IsSystemControlError(err) {
			fmt.Println("stream service cycle failed,", err)
			return
		}

		// for "some" system control error,
		// logical accept and retry stream service.
		if IsSystemControlRtmpClose(err) {
			fmt.Println("control message(close) accept, retry stream service.")
			continue
		}

		// for other system control message, fatal error.
		fmt.Println("control message reject as error")
		return
	}

	return
}
func (r *SrsClient) stream_service_cycle() (err error) {
	var client_type string
	if client_type, r.req.Stream, err = r.rtmp.IdentifyClient(r.res); err != nil {
		return
	}
	fmt.Printf("identify client success, type=%v, stream=%v\n", client_type, r.req.Stream)

	// set chunk size to larger.
	// TODO: FIXME: implements it.

	// find a source to serve.
	source := FindSrsSource(r.req)
	fmt.Println("discovery source", r.req.StreamUrl())

	// check publish available.
	// TODO: FIXME: implements it.

	// enable gop cache if requires
	// TODO: FIXME: implements it.

	switch client_type {
	case rtmp.CLIENT_TYPE_Play:
		if err = r.rtmp.StartPlay(r.res.StreamId()); err != nil {
			return
		}
		fmt.Println("start play stream")

		// on_play
		// TODO: FIXME: implements it.

		err = r.playing(source)

		// on_stop
		// TODO: FIXME: implements it.

		return err
	case rtmp.CLIENT_TYPE_FMLEPublish:
		if err = r.rtmp.StartFMLEPublish(r.res.StreamId()); err != nil {
			return
		}
		fmt.Println("start FMLE publish stream")

		// on_publish
		// TODO: FIXME: implements it.

		err = r.fmle_publishing(source)

		// on_unpublish
		// TODO: FIXME: implements it.
		return err
	case rtmp.CLIENT_TYPE_FlashPublish:
		if err = r.rtmp.StartFlashPublish(r.res.StreamId()); err != nil {
			return
		}
		fmt.Println("start flash publish stream")

		// on_publish
		// TODO: FIXME: implements it.

		err = r.flash_publishing(source)

		// on_unpublish
		// TODO: FIXME: implements it.

		return err
	}

	return
}

func (r *SrsClient) playing(source *SrsSource) (err error) {
	defer func() {
		if r.consumer == nil {
			return
		}

		if e := r.consumer.Close(); e != nil {
			if err == nil {
				err = e
			} else {
				fmt.Println("ignore the close error:", e)
			}
		}
		r.consumer = nil
	} ()

	// refer check
	// TODO: FIXME: implements it.

	r.consumer = source.CreateConsumer()

	// SrsPithyPrint
	// TODO: FIXME: implements it.
	for {
		r.conn.SetReadDeadline(time.Now().Add(time.Millisecond * SRS_PULSE_TIMEOUT_MS))

		// read from client.
		var msg *rtmp.Message
		if msg, err = r.rtmp.Protocol().RecvMessage(); err != nil {
			// if not tiemout error, return
			if neterr, ok := err.(net.Error); !ok || !neterr.Timeout() {
				return
			}
			// ignore the timeout error
			err = nil
		}

		if err = r.process_play_control_msg(msg); err != nil {
			return
		}

		// get messages from consumer.
		var consumer_has_message = true
		for consumer_has_message {
			select {
			case msg = <- r.consumer.msgs:
				// sendout messages
				if err = r.rtmp.Protocol().SendMessage(msg, uint32(r.res.StreamId())); err != nil {
					return
				}
			default:
				consumer_has_message = false
			}
		}
	}
	return
}
func (r *SrsClient) process_play_control_msg(msg *rtmp.Message) (err error) {
	// ignore all empty message.
	if msg == nil {
		return
	}

	if !msg.Header.IsAmf0Command() && !msg.Header.IsAmf3Command() {
		return
	}

	var pkt interface {}
	if pkt, err = r.rtmp.Protocol().DecodeMessage(msg); err != nil {
		return
	}

	if _, ok := pkt.(*rtmp.CloseStreamPacket); ok {
		return SrsError{code:ERROR_CONTROL_RTMP_CLOSE, desc:"system control message: rtmp close stream"}
	}

	// pause
	// TODO: FIXME: implements it
	return
}

func (r *SrsClient) fmle_publishing(source *SrsSource) (err error) {
	// refer check
	// TODO: FIXME: implements it.

	// notify the hls to prepare when publish start.
	// TODO: FIXME: implements it.

	for {
		// read from client.
		var msg *rtmp.Message
		if msg, err = r.rtmp.Protocol().RecvMessage(); err != nil {
			return
		}

		// process UnPublish event.
		if msg.Header.IsAmf0Command() || msg.Header.IsAmf3Command() {
			var pkt interface {}
			if pkt, err = r.rtmp.Protocol().DecodeMessage(msg); err != nil {
				return
			}

			if _, ok := pkt.(*rtmp.FMLEStartPacket); ok {
				fmt.Println("FMLE publish finished.")
				return
			}
			continue
		}

		if err = r.process_publish_message(source, msg); err != nil {
			return
		}
	}
	return
}
func (r *SrsClient) flash_publishing(source *SrsSource) (err error) {
	// refer check
	// TODO: FIXME: implements it.

	// notify the hls to prepare when publish start.
	// TODO: FIXME: implements it.

	for {
		// read from client.
		var msg *rtmp.Message
		if msg, err = r.rtmp.Protocol().RecvMessage(); err != nil {
			return
		}

		// process UnPublish event.
		if msg.Header.IsAmf0Command() || msg.Header.IsAmf3Command() {
			fmt.Println("flash publish finished.")
			return
		}

		if err = r.process_publish_message(source, msg); err != nil {
			return
		}
	}
	return
}
func (r *SrsClient) process_publish_message(source *SrsSource, msg *rtmp.Message) (err error) {
	// process audio packet
	if msg.Header.IsAudio() {
		if err = source.OnAudio(msg); err != nil {
			return
		}
	}

	// process video packet
	if msg.Header.IsVideo() {
		if err = source.OnVideo(msg); err != nil {
			return
		}
	}

	// process onMetaData
	// TODO: FIXME: implements it.
	return
}

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
	"io"
	"github.com/winlinvip/go.rtmp/rtmp"
)

// default stream id for response the createStream request.
const SRS_DEFAULT_SID = 1

/**
* the response info for srs.
 */
type SrsResponse struct {
	stream_id uint32
}
func NewSrsResponse() (*SrsResponse) {
	r := &SrsResponse{}
	r.stream_id = SRS_DEFAULT_SID
	return r
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
	id SrsLogId
}
func NewSrsClient(conn *net.TCPConn) (r *SrsClient, err error) {
	r = &SrsClient{}
	r.conn = conn
	r.res = NewSrsResponse()
	r.id = SrsGenerateId()

	if r.rtmp, err = rtmp.NewServer(conn); err != nil {
		return
	}
	r.req = rtmp.NewRequest()
	return
}

// interface for Log
func (r *SrsClient) GetId() (SrsLogId) {
	return r.id
}
func (r *SrsClient) GetTag() (SrsLogTag) {
	return "client"
}

func (r *SrsClient) do_cycle() (err error) {
	defer func(r *SrsClient) {
		if rc := recover(); rc != nil {
			SrsWarn(r, r, "ignore panic from serve client, err=%v, rc=%v", err, rc)
			return
		}

		// ignore the normally closed
		if err == nil {
			return
		}
		SrsTrace(r, r, "client cycle completed, err=%v", err)
	}(r)

	SrsTrace(r, r, "start serve client=%v", r.conn.RemoteAddr())

	r.rtmp.Protocol().SetReadTimeout(SRS_RECV_TIMEOUT_MS)
	r.rtmp.Protocol().SetWriteTimeout(SRS_SEND_TIMEOUT_MS)

	if err = r.rtmp.Handshake(); err != nil {
		return
	}

	if err = r.rtmp.ConnectApp(r.req); err != nil {
		return
	}
	SrsTrace(r, r, "request, tcUrl=%v(vhost=%v, app=%v), AMF%v, pageUrl=%v, swfUrl=%v",
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
	SrsTrace(r, r, "set window ack size to %v", ack_size)

	bandwidth, bw_type := uint32(2.5 * 1000 * 1000), byte(2)
	if err = r.rtmp.SetPeerBandwidth(bandwidth, bw_type); err != nil {
		return
	}
	SrsTrace(r, r, "set bandwidth to %v, type=%v", bandwidth, bw_type)

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
	SrsTrace(r, r, "response connect app success")

	if err = r.rtmp.CallOnBWDone(); err != nil {
		return
	}
	SrsTrace(r, r, "call client as onBWDone()")

	for {
		err = r.stream_service_cycle()

		// stream service must terminated with error, never success.
		if err == nil {
			SrsTrace(r, r, "stream service complete success, re-identify it")
			continue
		}

		// when not system control error, fatal error, return.
		if !IsSystemControlError(err) {
			if err == io.EOF {
				SrsTrace(r, r, "client gracefully close the peer")
				err = nil
				return
			}
			SrsWarn(r, r, "stream service cycle failed, err=%v", err)
			return
		}

		// for "some" system control error,
		// logical accept and retry stream service.
		if IsSystemControlRtmpClose(err) {
			SrsWarn(r, r, "control message(close) accept, retry stream service.")

			// set timeout to a larger value, for user paused.
			r.rtmp.Protocol().SetReadTimeout(SRS_PAUSED_RECV_TIMEOUT_MS)
			r.rtmp.Protocol().SetWriteTimeout(SRS_PAUSED_SEND_TIMEOUT_MS)

			continue
		}

		// for other system control message, fatal error.
		SrsTrace(r, r, "control message reject as error, err=%v", err)
		return
	}

	return
}
func (r *SrsClient) stream_service_cycle() (err error) {
	var client_type string
	if client_type, r.req.Stream, err = r.rtmp.IdentifyClient(r.res.stream_id); err != nil {
		return
	}
	SrsTrace(r, r, "identify client success, type=%v, stream=%v", client_type, r.req.Stream)

	// client is identified, set the timeout to service timeout.
	r.rtmp.Protocol().SetReadTimeout(SRS_RECV_TIMEOUT_MS)
	r.rtmp.Protocol().SetWriteTimeout(SRS_SEND_TIMEOUT_MS)

	// set chunk size to larger.
	// TODO: FIXME: implements it.

	// find a source to serve.
	source := FindSrsSource(r.req)
	SrsTrace(r, r, "discovery source by url %v", r.req.StreamUrl())

	// check publish available.
	// TODO: FIXME: implements it.

	// enable gop cache if requires
	// TODO: FIXME: implements it.

	switch client_type {
	case rtmp.CLIENT_TYPE_Play:
		if err = r.rtmp.StartPlay(r.res.stream_id); err != nil {
			return
		}
		SrsTrace(r, r, "start play stream")

		// on_play
		// TODO: FIXME: implements it.

		err = r.playing(source)

		// on_stop
		// TODO: FIXME: implements it.

		return err
	case rtmp.CLIENT_TYPE_FMLEPublish:
		if err = r.rtmp.StartFMLEPublish(r.res.stream_id); err != nil {
			return
		}
		SrsTrace(r, r, "start FMLE publish stream")

		// on_publish
		// TODO: FIXME: implements it.

		err = r.fmle_publishing(source)

		// on_unpublish
		// TODO: FIXME: implements it.
		return err
	case rtmp.CLIENT_TYPE_FlashPublish:
		if err = r.rtmp.StartFlashPublish(r.res.stream_id); err != nil {
			return
		}
		SrsTrace(r, r, "start flash publish stream")

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
				SrsTrace(r, r, "ignore the close err=%v", e)
			}
		}
		r.consumer = nil
	} ()

	// refer check
	// TODO: FIXME: implements it.

	r.consumer = source.CreateConsumer()

	r.rtmp.Protocol().SetReadTimeout(SRS_PULSE_TIMEOUT_MS)

	// SrsPithyPrint
	// TODO: FIXME: implements it.

	for {
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
		msgs := r.consumer.Messages()
		for i := 0; i < len(msgs); i++ {
			msg := msgs[i]
			if msg == nil {
				break
			}
			// sendout messages
			if err = r.rtmp.Protocol().SendMessage(msg, r.res.stream_id); err != nil {
				return
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
				SrsTrace(r, r, "FMLE publish finished.")
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
			SrsTrace(r, r, "flash publish finished.")
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

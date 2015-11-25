package main

import (
	"io"
	"net"
	"runtime"

	"github.com/Alienero/IamServer/rtmp"
	"github.com/Alienero/IamServer/source"

	"github.com/golang/glog"
)

// default stream id for response the createStream request.
const SRS_DEFAULT_SID = 1

// the response info for srs.
type SrsResponse struct {
	stream_id uint32
}

func NewSrsResponse() *SrsResponse {
	r := &SrsResponse{}
	r.stream_id = SRS_DEFAULT_SID
	return r
}

// the client provides the main logic control for RTMP clients.
type SrsClient struct {
	conn     *net.TCPConn
	rtmp     rtmp.Server
	req      *rtmp.Request
	res      *SrsResponse
	consumer *SrsConsumer
	id       uint64
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

func (r *SrsClient) do_cycle() (err error) {
	defer func() {
		r.conn.Close()
		glog.Info("will Destroy")
		// destroy the protocol stack.
		r.rtmp.Destroy() // here has dead lock.
		glog.Info("here")
		if rc := recover(); rc != nil {
			buff := make([]byte, 4096)
			runtime.Stack(buff, false)
			glog.Errorf("ignore panic from serve client, err=%v, rc=%v,stack:%v", err, rc, string(buff))
			return
		}

		// ignore the normally closed
		if err == nil {
			glog.Info("OK client close.")
			return
		}
		glog.Errorf("client cycle completed, err=%v", err)
	}()

	glog.Infof("start serve client=%v", r.conn.RemoteAddr())

	if err = r.rtmp.Handshake(); err != nil {
		return
	}

	if err = r.rtmp.ConnectApp(r.req); err != nil {
		return
	}
	glog.Infof("request, tcUrl=%v(vhost=%v, app=%v), AMF%v, pageUrl=%v, swfUrl=%v",
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
	glog.Infof("set window ack size to %v", ack_size)

	bandwidth, bw_type := uint32(2.5*1000*1000), byte(2)
	if err = r.rtmp.SetPeerBandwidth(bandwidth, bw_type); err != nil {
		return
	}
	glog.Infof("set bandwidth to %v, type=%v", bandwidth, bw_type)

	// do bandwidth test if connect to the vhost which is for bandwidth check.
	// TODO: FIXME: implements it

	extra_data := []map[string]string{
		{"srs_sig": RTMP_SIG_SRS_KEY},
		{"srs_server": RTMP_SIG_SRS_KEY + " " + RTMP_SIG_SRS_VERSION + " (" + RTMP_SIG_SRS_URL_SHORT + ")"},
		{"srs_license": RTMP_SIG_SRS_LICENSE},
		{"srs_role": RTMP_SIG_SRS_ROLE},
		{"srs_url": RTMP_SIG_SRS_URL},
		{"srs_version": RTMP_SIG_SRS_VERSION},
		{"srs_site": RTMP_SIG_SRS_WEB},
		{"srs_email": RTMP_SIG_SRS_EMAIL},
		{"srs_copyright": RTMP_SIG_SRS_COPYRIGHT},
		{"srs_primary_authors": RTMP_SIG_SRS_PRIMARY_AUTHROS},
	}
	if err = r.rtmp.ReponseConnectApp(r.req, "", extra_data); err != nil {
		return
	}
	glog.Infof("response connect app success")

	if err = r.rtmp.CallOnBWDone(); err != nil {
		return
	}
	glog.Infof("call client as onBWDone()")
	defer func() { glog.Info("Connection is closed!") }()
	for {
		err = r.stream_service_cycle()

		// stream service must terminated with error, never success.
		if err == nil {
			glog.Infof("stream service complete success, re-identify it")
			continue
		}

		// when not system control error, fatal error, return.
		if !IsSystemControlError(err) {
			if err == io.EOF {
				glog.Infof("client gracefully close the peer")
				err = nil
				return
			}
			glog.Warningf("stream service cycle failed, err=%v", err)
			return
		}

		// for "some" system control error,
		// logical accept and retry stream service.
		if IsSystemControlRtmpClose(err) {
			glog.Warningf("control message(close) accept, retry stream service.")
			continue
		}

		// for other system control message, fatal error.
		glog.Infof("control message reject as error, err=%v", err)
		return
	}

	return
}

func (r *SrsClient) stream_service_cycle() (err error) {
	var client_type string
	if client_type, r.req.Stream, err = r.rtmp.IdentifyClient(r.res.stream_id); err != nil {
		return
	}
	glog.Infof("identify client success, type=%v, stream=%v", client_type, r.req.Stream)

	// set chunk size to larger.
	// TODO: FIXME: implements it.

	// set a source to serve.
	key := "/" + r.req.App + "/" + r.req.Stream
	s, err := source.Sources.Set(key)
	if err != nil {
		return err
	}
	s.SetFlvHead()
	defer func() {
		source.Sources.Delete(key)
		s.Close()
		glog.Info("free sources.")
	}()

	glog.Infof("discovery source by url %v", r.req.StreamUrl())

	// check publish available.
	// TODO: FIXME: implements it.

	// enable gop cache if requires
	// TODO: FIXME: implements it.

	switch client_type {
	case rtmp.CLIENT_TYPE_Play:
		// just return ,not support.
		return err
	case rtmp.CLIENT_TYPE_FMLEPublish:
		if err = r.rtmp.StartFMLEPublish(r.res.stream_id); err != nil {
			return
		}
		glog.Info("start FMLE publish stream")

		// on_publish
		// TODO: FIXME: implements it.

		err = r.fmle_publishing(s)

		// on_unpublish
		// TODO: FIXME: implements it.
		return err
	case rtmp.CLIENT_TYPE_FlashPublish:
		// just return, not spport.
		return err
	}

	return
}

func (r *SrsClient) fmle_publishing(s *source.Sourcer) (err error) {
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
			var pkt interface{}
			if pkt, err = r.rtmp.Protocol().DecodeMessage(msg); err != nil {
				return
			}

			if _, ok := pkt.(*rtmp.FMLEStartPacket); ok {
				glog.Info("FMLE publish finished.")
				return
			}
			continue
		}
		if msg.Header.Timestamp == 0 && msg.Header.IsAmf0Data() {
			glog.Info("set meta data")
			if err = s.SetMeta(msg); err != nil {
				return
			}
			glog.Info("set meta data done.")
			continue
		}

		if err = r.process_publish_message(s, msg); err != nil {
			return
		}
	}
	return
}

func (r *SrsClient) process_publish_message(s *source.Sourcer, msg *rtmp.Message) (err error) {
	// glog.Info("handle msg")
	s.HandleMsg(msg)
	// glog.Info("handle msg done")
	return
}

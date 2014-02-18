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
	"fmt"
	"net"
	"github.com/winlinvip/go.rtmp/rtmp"
)

// current release version
const RTMP_SIG_SRS_VERSION = "0.1.0"
// = server = info.
const RTMP_SIG_SRS_KEY = "go.srs"
const RTMP_SIG_SRS_ROLE = "stream server"
const RTMP_SIG_SRS_NAME = RTMP_SIG_SRS_KEY+"(simple = rtmp = server)"
const RTMP_SIG_SRS_URL_SHORT = "github.com/winlinvip/go.srs"
const RTMP_SIG_SRS_URL = "https://"+RTMP_SIG_SRS_URL_SHORT
const RTMP_SIG_SRS_WEB = "http://blog.csdn.net/win_lin"
const RTMP_SIG_SRS_EMAIL = "winlin@vip.126.com"
const RTMP_SIG_SRS_LICENSE = "The MIT License (MIT)"
const RTMP_SIG_SRS_COPYRIGHT = "Copyright (c) 2014 winlin"
const RTMP_SIG_SRS_PRIMARY_AUTHROS = "winlin"

/**
* the client provides the main logic control for RTMP clients.
*/
type SrsClient struct {
	conn *net.TCPConn
	rtmp rtmp.RtmpServer
	req *rtmp.RtmpRequest
}
func NewSrsClient(conn *net.TCPConn) (r *SrsClient, err error) {
	r = &SrsClient{}
	r.conn = conn

	if r.rtmp, err = rtmp.NewRtmpServer(conn); err != nil {
		return
	}
	r.req = rtmp.NewRtmpRequest()
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

	// TODO: FIXME: implements it
	return
}

func main() {
    fmt.Println("SRS(simple-rtmp-server) written by google go language.")
    fmt.Println("RTMP Protocol Stack: ", rtmp.Version)

	addr, err := net.ResolveTCPAddr("tcp4", ":1935")
	if err != nil {
		fmt.Println("error:", err)
		return;
	}

	var listener *net.TCPListener
	listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		fmt.Println("error:", err)
		return;
	}
	defer listener.Close()

	for {
		fmt.Println("listener ready to accept client")
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("error:", err)
			return;
		}

		serve := func(conn *net.TCPConn) {
			defer conn.Close()

			fmt.Println("get client:", conn.RemoteAddr())

			var client *SrsClient
			if client, err = NewSrsClient(conn); err != nil {
				fmt.Println("error:", err)
				return
			}

			if err = client.do_cycle(); err != nil {
				fmt.Println("serve client error:", err)
			} else {
				fmt.Println("serve client completed")
			}
		}
		go serve(conn)
	}
}

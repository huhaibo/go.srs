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
	"github.com/winlinvip/go.rtmp/rtmp"
)

type SrsServer struct {
	id SrsLogId
}
func NewSrsServer() (*SrsServer) {
	r := &SrsServer{}
	r.id = SrsGenerateId()
	return r
}

// interface for Log
func (r *SrsServer) GetId() (SrsLogId) {
	return r.id
}
func (r *SrsServer) GetTag() (SrsLogTag) {
	return "client"
}

func (r *SrsServer) PrintInfo() {
	SrsTrace(r, r, "SRS(simple-rtmp-server) written by google go language.")
	SrsTrace(r, r, "RTMP Protocol Stack:  %v", rtmp.Version)
}

func (r *SrsServer) Serve() {
	addr, err := net.ResolveTCPAddr("tcp4", ":1935")
	if err != nil {
		SrsFatal(r, r, "resolve listen address failed, err=%v", err)
		return;
	}

	var listener *net.TCPListener
	listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		SrsFatal(r, r, "listen failed, err=%v", err)
		return;
	}
	defer listener.Close()

	for {
		SrsVerbose(r, r, "listener ready to accept client")
		conn, err := listener.AcceptTCP()
		if err != nil {
			SrsFatal(r, r, "accept client failed, err=%v", err)
			return;
		}

		serve := func(conn *net.TCPConn) {
			defer conn.Close()

			var client *SrsClient
			if client, err = NewSrsClient(conn); err != nil {
				SrsFatal(r, r, "create client failed, err=%v", err)
				return
			}

			err = client.do_cycle()
		}
		go serve(conn)
	}
}

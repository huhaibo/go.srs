package main

import (
	"fmt"
	"net"

	"github.com/Alienero/IamServer/rtmp"

	"github.com/golang/glog"
)

type SrsServer struct {
	id uint64
}

func NewSrsServer() *SrsServer {
	r := &SrsServer{}
	r.id = SrsGenerateId()
	return r
}

func (r *SrsServer) PrintInfo() {
	glog.Infof("RTMP Protocol Stack:  %v", rtmp.Version)
}

func (r *SrsServer) Serve() error {
	addr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		glog.Errorf("resolve listen address failed, err=%v", err)
		return fmt.Errorf("resolve listen address failed, err=%v", err)
	}

	var listener *net.TCPListener
	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		glog.Errorf("listen failed, err=%v", err)
		return fmt.Errorf("listen failed, err=%v", err)
	}
	defer listener.Close()
	for {
		glog.Info("listener ready to accept client")
		conn, err := listener.AcceptTCP()
		if err != nil {
			glog.Errorf("accept client failed, err=%v", err)
			return fmt.Errorf("accept client failed, err=%v", err)
		}
		glog.Info("TCP Connected")

		go r.serve(conn)
	}
}

func (r *SrsServer) serve(conn *net.TCPConn) {
	var (
		client *SrsClient
		err    error
	)
	if client, err = NewSrsClient(conn); err != nil {
		glog.Errorf("create client failed, err=%v", err)
		return
	}

	if err = client.do_cycle(); err != nil {
		glog.Errorf("do cycle err=%v", err)
	}
}

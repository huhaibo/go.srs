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

type SrsError struct {
	code int
	desc string
}
func (err SrsError) Error() string {
	return fmt.Sprintf("code=%v: %s", err.code, err.desc)
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

	for {
		fmt.Println("listener ready to accept client")
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("error:", err)
			return;
		}

		SimpleHandshake := func(conn *net.TCPConn) (err error) {
			c0c1 := make([]byte, 1537)
			// TODO: FIXME: read in block mode.
			nsize, err := conn.Read(c0c1)
			if err != nil {
				return
			}
			fmt.Println("read c0c1, size=", nsize)

			s0s1s2 := make([]byte, 3073)
			copy(s0s1s2[0:1537], c0c1)
			// TODO: FIXME: write in block mode.
			nsize, err = conn.Write(s0s1s2)
			if err != nil {
				return
			}
			fmt.Println("write s0s1s2, size=", nsize)

			c2 := make([]byte, 1536)
			nsize, err = conn.Read(c2)
			if err != nil {
				return
			}
			fmt.Println("read c2, size=", nsize)

			return
		}

		ReadBasicHeader := func(conn *net.TCPConn) (format byte, cid int, bh_size int, err error) {
			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				return
			}

			format = buf[0]
			cid = int(format) & 0x3f
			format = (format >> 6) & 0x03
			bh_size = 1

			if cid > 1 {
				return
			}

			if cid == 0 {
				_, err = conn.Read(buf)
				if err != nil {
					return
				}
				cid = 64
				cid += int(buf[0])
				bh_size = 2
			} else if cid == 1 {
				buf = make([]byte, 2)
				_, err = conn.Read(buf)
				if err != nil {
					return
				}
				cid = 64
				cid += int(buf[0])
				cid += int(buf[1]) * 256
				bh_size = 3
			} else {
				err = SrsError{}
			}

			return
		}

		do_serve :=func(conn *net.TCPConn) (err error) {
			err = SimpleHandshake(conn)
			if err != nil {
				return
			}

			format, cid, bhsize, err := ReadBasicHeader(conn)
			if err != nil {
				return
			}
			fmt.Printf("fmt=%v, cid=%v, bhsize=%v\n", format, cid, bhsize)

			return
		}

		serve := func(conn *net.TCPConn) {
			defer conn.Close()

			fmt.Println("get client:", conn.RemoteAddr())
			err := do_serve(conn)
			if err != nil {
				fmt.Println("error:", err)
				return
			}

			fmt.Println("serve client completed")
		}
		go serve(conn)
	}
}

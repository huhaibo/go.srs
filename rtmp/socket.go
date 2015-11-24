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

package rtmp

import (
	"net"
	"fmt"
)

// socket to read or write data.
type Socket struct {
	conn *net.TCPConn
	recv_bytes uint64
	send_bytes uint64
}
func NewSocket(conn *net.TCPConn) (*Socket) {
	r := &Socket{}
	r.conn = conn
	return r
}

func (r *Socket) RecvBytes() (uint64) {
	return r.recv_bytes
}

func (r *Socket) SendBytes() (uint64) {
	return r.send_bytes
}

func (r *Socket) Read(b []byte) (n int, err error) {
	if n, err = r.conn.Read(b); err != nil {
		return
	}

	if n == 0 {
		if err == nil {
			err = Error{code:ERROR_SOCKET_CLOSED, desc:"read peer closed gracefully"}
		}
		return
	}

	if n < 0 {
		if err == nil {
			err = Error{code:ERROR_SOCKET_READ, desc:"read data failed"}
		}
		return
	}

	if n > 0 {
		r.recv_bytes += uint64(n)
	}

	return
}

func (r *Socket) Write(b []byte) (n int, err error) {
	for n < len(b) {
		var nb_written int
		if nb_written, err = r.conn.Write(b[n:]); err != nil {
			return
		}

		if nb_written == 0 {
			if err == nil {
				err = Error{code:ERROR_SOCKET_CLOSED, desc:"write peer closed gracefully"}
			}
			return
		}

		if nb_written < 0 {
			if err == nil {
				err = Error{code:ERROR_SOCKET_WRITE, desc:"write data failed"}
			}
			return
		}

		r.send_bytes += uint64(nb_written)
		n += nb_written

		if n < len(b) {
			// TODO: FIXME: remove following
			fmt.Printf("write partially, len(b)=%v, n=%v, nb_written=%v\n", len(b), n, nb_written)
		}
	}

	return
}

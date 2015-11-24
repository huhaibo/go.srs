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
	"io"
	"math/rand"
)

func (r *protocol) handshake_read_c0c1() (err error) {
	var handshake *Handshake = r.handshake

	if handshake.c0c1 == nil {
		handshake.c0c1 = make([]byte, 1537)
		if _, err = io.ReadFull(r.conn, handshake.c0c1); err != nil {
			return
		}
	}

	return
}
func (r *protocol) handshake_make_s0s1s2() (err error) {
	var handshake *Handshake = r.handshake

	if handshake.s0s1s2 == nil {
		handshake.s0s1s2 = make([]byte, 3073)
	}

	return
}
func (r *protocol) handshake_read_c2() (err error) {
	var handshake *Handshake = r.handshake

	if handshake.c2 == nil {
		handshake.c2 = make([]byte, 1536)
		if _, err = io.ReadFull(r.conn, handshake.c2); err != nil {
			return
		}
	}

	return
}

func (r *protocol) SimpleHandshake2Client() (err error) {
	var handshake *Handshake = r.handshake

	// read the c0c1 from connection if not read yet
	if err = r.handshake_read_c0c1(); err != nil {
		return
	}

	// plain text required.
	if handshake.c0c1[0] != 0x03 {
		err = Error{code:ERROR_RTMP_PLAIN_REQUIRED, desc:"only support rtmp plain text"}
		return
	}

	// genereate the s0s1s2, alloc the bytes
	if err = r.handshake_make_s0s1s2(); err != nil {
		return
	}

	// for simple handshake, fill the s0s1s2 with random data
	for i, _ := range handshake.s0s1s2 {
		handshake.s0s1s2[i] = byte(rand.Int())
	}
	// plain text required.
	handshake.s0s1s2[0] = 0x03

	// for simple handshake, directly write the s0s1s2
	if _, err = r.conn.Write(handshake.s0s1s2); err != nil {
		return
	}

	// read the c2 from connection if not read yet
	if err = r.handshake_read_c2(); err != nil {
		return
	}

	// start messages input/outout goroutines
	r.start_message_pump_goroutines()

	return
}

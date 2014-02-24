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

// the following is the timeout for rtmp protocol,
// to avoid death connection.

// when got a messae header, there must be some data,
// increase recv timeout to got an entire message.
const SRS_MIN_RECV_TIMEOUT_MS = 60*1000

// the timeout to wait for client control message,
// if timeout, we generally ignore and send the data to client,
// generally, it's the pulse time for data seding.
const SRS_PULSE_TIMEOUT_MS = 200

// the pprof timeout, the ping message recv/send timeout
const SRS_PPROF_TIMEOUT_MS = 10*60*1000
const SRS_PPROF_PULSE_MS = 800
const SRS_PPROF_VHOST = "pprof"

// the timeout to wait client data,
// if timeout, close the connection.
const SRS_SEND_TIMEOUT_MS = 30*1000

// the timeout to send data to client,
// if timeout, close the connection.
const SRS_RECV_TIMEOUT_MS = 30*1000

// the timeout to wait client data, when client paused
// if timeout, close the connection.
const SRS_PAUSED_SEND_TIMEOUT_MS = 30*60*1000

// the timeout to send data to client, when client paused
// if timeout, close the connection.
const SRS_PAUSED_RECV_TIMEOUT_MS = 30*60*1000

// when stream is busy, for example, streaming is already
// publishing, when a new client to request to publish,
// sleep a while and close the connection.
const SRS_STREAM_BUSY_SLEEP_MS = 3*1000

// when error, forwarder sleep for a while and retry.
const SRS_FORWARDER_SLEEP_MS = 3*1000

// when error, encoder sleep for a while and retry.
const SRS_ENCODER_SLEEP_MS = 3*1000

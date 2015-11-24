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

import "fmt"

const ERROR_SUCCESS = 0

const ERROR_GO_REFLECT_PTR_REQUIRES = 100
const ERROR_GO_REFLECT_NEVER_NIL = 101
const ERROR_GO_REFLECT_CAN_SET = 102
const ERROR_GO_AMF0_NIL_PROPERTY = 103
const ERROR_GO_RTMP_NOT_SUPPORT_MSG = 104
const ERROR_GO_PROTOCOL_DESTROYED = 105

const ERROR_SOCKET_CREATE = 200
const ERROR_SOCKET_SETREUSE = 201
const ERROR_SOCKET_BIND = 202
const ERROR_SOCKET_LISTEN = 203
const ERROR_SOCKET_CLOSED = 204
const ERROR_SOCKET_GET_PEER_NAME = 205
const ERROR_SOCKET_GET_PEER_IP = 206
const ERROR_SOCKET_READ = 207
const ERROR_SOCKET_READ_FULLY = 208
const ERROR_SOCKET_WRITE = 209
const ERROR_SOCKET_WAIT = 210
const ERROR_SOCKET_TIMEOUT = 211
const ERROR_SOCKET_GET_LOCAL_IP = 212

const ERROR_RTMP_PLAIN_REQUIRED = 300
const ERROR_RTMP_CHUNK_START = 301
const ERROR_RTMP_MSG_INVLIAD_SIZE = 302
const ERROR_RTMP_AMF0_DECODE = 303
const ERROR_RTMP_AMF0_INVALID = 304
const ERROR_RTMP_REQ_CONNECT = 305
const ERROR_RTMP_REQ_TCURL = 306
const ERROR_RTMP_MESSAGE_DECODE = 307
const ERROR_RTMP_MESSAGE_ENCODE = 308
const ERROR_RTMP_AMF0_ENCODE = 309
const ERROR_RTMP_CHUNK_SIZE = 310
const ERROR_RTMP_TRY_SIMPLE_HS = 311
const ERROR_RTMP_CH_SCHEMA = 312
const ERROR_RTMP_PACKET_SIZE = 313
const ERROR_RTMP_VHOST_NOT_FOUND = 314
const ERROR_RTMP_ACCESS_DENIED = 315
const ERROR_RTMP_HANDSHAKE = 316
const ERROR_RTMP_NO_REQUEST = 317

const ERROR_SYSTEM_STREAM_INIT = 400
const ERROR_SYSTEM_PACKET_INVALID = 401
const ERROR_SYSTEM_CLIENT_INVALID = 402
const ERROR_SYSTEM_ASSERT_FAILED = 403
const ERROR_SYSTEM_SIZE_NEGATIVE = 404
const ERROR_SYSTEM_CONFIG_INVALID = 405
const ERROR_SYSTEM_CONFIG_DIRECTIVE = 406
const ERROR_SYSTEM_CONFIG_BLOCK_START = 407
const ERROR_SYSTEM_CONFIG_BLOCK_END = 408
const ERROR_SYSTEM_CONFIG_EOF = 409
const ERROR_SYSTEM_STREAM_BUSY = 410
const ERROR_SYSTEM_IP_INVALID = 411
const ERROR_SYSTEM_FORWARD_LOOP = 412
const ERROR_SYSTEM_WAITPID = 413
const ERROR_SYSTEM_BANDWIDTH_KEY = 414
const ERROR_SYSTEM_BANDWIDTH_DENIED = 415

// see librtmp.
// failed when open ssl create the dh
const ERROR_OpenSslCreateDH = 500
// failed when open ssl create the Private key.
const ERROR_OpenSslCreateP = 501
// when open ssl create G.
const ERROR_OpenSslCreateG = 502
// when open ssl parse P1024
const ERROR_OpenSslParseP1024 = 503
// when open ssl set G
const ERROR_OpenSslSetG = 504
// when open ssl generate DHKeys
const ERROR_OpenSslGenerateDHKeys = 505
// when open ssl share key already computed.
const ERROR_OpenSslShareKeyComputed = 506
// when open ssl get shared key size.
const ERROR_OpenSslGetSharedKeySize = 507
// when open ssl get peer public key.
const ERROR_OpenSslGetPeerPublicKey = 508
// when open ssl compute shared key.
const ERROR_OpenSslComputeSharedKey = 509
// when open ssl is invalid DH state.
const ERROR_OpenSslInvalidDHState = 510
// when open ssl copy key
const ERROR_OpenSslCopyKey = 511
// when open ssl sha256 digest key invalid size.
const ERROR_OpenSslSha256DigestSize = 512

const ERROR_HLS_METADATA = 600
const ERROR_HLS_DECODE_ERROR = 601
const ERROR_HLS_CREATE_DIR = 602
const ERROR_HLS_OPEN_FAILED = 603
const ERROR_HLS_WRITE_FAILED = 604
const ERROR_HLS_AAC_FRAME_LENGTH = 605
const ERROR_HLS_AVC_SAMPLE_SIZE = 606

const ERROR_ENCODER_VCODEC = 700
const ERROR_ENCODER_OUTPUT = 701
const ERROR_ENCODER_ACHANNELS = 702
const ERROR_ENCODER_ASAMPLE_RATE = 703
const ERROR_ENCODER_ABITRATE = 704
const ERROR_ENCODER_ACODEC = 705
const ERROR_ENCODER_VPRESET = 706
const ERROR_ENCODER_VPROFILE = 707
const ERROR_ENCODER_VTHREADS = 708
const ERROR_ENCODER_VHEIGHT = 709
const ERROR_ENCODER_VWIDTH = 710
const ERROR_ENCODER_VFPS = 711
const ERROR_ENCODER_VBITRATE = 712
const ERROR_ENCODER_FORK = 713
const ERROR_ENCODER_LOOP = 714
const ERROR_ENCODER_OPEN = 715
const ERROR_ENCODER_DUP2 = 716

const ERROR_HTTP_PARSE_URI = 800
const ERROR_HTTP_DATA_INVLIAD = 801
const ERROR_HTTP_PARSE_HEADER = 802

type Error struct {
	code int
	desc string
}
func (err Error) Error() string {
	return fmt.Sprintf("rtmp error code=%v: %s", err.code, err.desc)
}

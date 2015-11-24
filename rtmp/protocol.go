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
	"math"
	"reflect"
	// "runtime"
)

/**
* the handshake data, 6146B = 6KB,
* store in the protocol and never delete it for every connection.
 */
type Handshake struct {
	c0c1   []byte // 1537B
	s0s1s2 []byte // 3073B
	c2     []byte // 1536B
}

type AckWindowSize struct {
	ack_window_size uint32
	acked_size      uint64
}

// should ack the read, ack to peer
func (r *AckWindowSize) ShouldAckRead(n uint64) bool {
	if r.ack_window_size <= 0 {
		return false
	}

	return n-uint64(r.acked_size) > uint64(r.ack_window_size)
}

/**
* the protocol provides the rtmp-message-protocol services,
* to recv RTMP message from RTMP chunk stream,
* and to send out RTMP message over RTMP chunk stream.
 */
type protocol struct {
	// handshake
	handshake *Handshake
	// peer in/out
	// the underlayer tcp connection, to read/write bytes from/to.
	conn *Socket
	/**
	* requests sent out, used to build the response.
	* key: a float64 indicates the transactionId
	* value: a string indicates the request command name
	 */
	requests map[float64]string
	// peer in
	chunkStreams map[int]*ChunkStream
	// the bytes read from underlayer tcp connection,
	// used for parse to RTMP message or packets.
	buffer *Buffer
	// input chunk stream chunk size.
	inChunkSize uint32
	// the acked size
	inAckSize AckWindowSize
	// peer out
	// output chunk stream chunk size.
	outChunkSize uint32
	// bytes cache, size is RTMP_MAX_FMT0_HEADER_SIZE
	outHeaderFmt0 *Buffer
	// bytes cache, size is RTMP_MAX_FMT3_HEADER_SIZE
	outHeaderFmt3 *Buffer
	// use channel to store the decoded message, or messages to encode,
	// for user can use select to determinate the event of message(incoming or outgoing)
	// message channel lock, to stop protocol
	// msg_in_lock  *sync.Mutex
	// msg_out_lock *sync.Mutex
	// input/output error
	msg_io_err error
	// message input queue, received message from connection.
	msg_in_queue chan *Message
	// message output queue, message to send over connection
	msg_out_queue chan *Message

	// close signal.
	closeChan chan struct{}
}

/**
* destroy the protocol stack, close channels, stop goroutines.
 */
func (r *protocol) Destroy() {
	// r.msg_in_lock.Lock()
	// r.msg_out_lock.Lock()
	// defer r.msg_out_lock.Unlock()
	// defer r.msg_in_lock.Unlock()
	r.closeChan <- struct{}{}

	if r.msg_io_err == nil {
		r.msg_io_err = Error{code: ERROR_GO_PROTOCOL_DESTROYED, desc: "protocol stack destroyed"}
	}

	// If recive an error, close in_queue should in recieve goroutine.
	// Let Recive method know conn get an error.
	// close(r.msg_in_queue)

	close(r.msg_out_queue)
	close(r.closeChan)
}

func (r *protocol) MessageInputChannel() chan *Message {
	return r.msg_in_queue
}

/**
* start pump messages, input/output goroutines:
* recv message from connection and put into msg_in_queue
* send messages in msg_out_queue over connection
 */
func (r *protocol) start_message_pump_goroutines() {
	go r.recv_msg_goroutine()
	go r.send_msg_goroutine()
}
func (r *protocol) recv_msg_goroutine() {
	for {

		var (
			msg *Message
			err error
		)

		if msg, err = r.recv_interlaced_message(); err != nil {
			goto check
		}

		if msg == nil {
			continue
		}

		if msg.ReceivedPayloadLength <= 0 || msg.Header.PayloadLength <= 0 {
			continue
		}

		err = r.on_recv_message(msg)

	check:
		if err != nil {
			println("Get an error", err.Error())
			r.msg_io_err = err
			// in there close in_queue channel. If not Recive method maybe hang.
			close(r.msg_in_queue)
			<-r.closeChan
			return
		} else {
			select {
			case <-r.closeChan:
				close(r.msg_in_queue)
				return
			case r.msg_in_queue <- msg:
				// continue loop.
			}
		}
	}
}

func (r *protocol) send_msg_goroutine() {
	for {
		if err := r.do_send_msg_goroutine_job(); err != nil {
			r.msg_io_err = err
			return
		}
	}
}

func (r *protocol) do_recv_msg_goroutine_job() (err error) {
	var msg *Message

	if msg, err = r.recv_interlaced_message(); err != nil {
		return
	}

	if msg == nil {
		return Error{code: ERROR_RTMP_MSG_INVLIAD_SIZE, desc: "msg length invliad"}
	}

	if msg.ReceivedPayloadLength <= 0 || msg.Header.PayloadLength <= 0 {
		return Error{code: ERROR_RTMP_MSG_INVLIAD_SIZE, desc: "msg length invliad"}
	}

	if err = r.on_recv_message(msg); err != nil {
		return
	}
	select {
	case <-r.closeChan:
		// pass
	case r.msg_in_queue <- msg:
		// pass
	}
	return
}
func (r *protocol) do_send_msg_goroutine_job() (err error) {
	var msg *Message
	var ok bool
	if msg, ok = <-r.msg_out_queue; !ok {
		err = Error{code: ERROR_GO_PROTOCOL_DESTROYED, desc: "protocol stack destroyed, cannot send"}
		return
	}

	// always write the header event payload is empty.
	msg.SentPayloadLength = -1
	for len(msg.Payload) > msg.SentPayloadLength {
		msg.SentPayloadLength = int(math.Max(0, float64(msg.SentPayloadLength)))

		// generate the header.
		var real_header []byte

		if msg.SentPayloadLength <= 0 {
			// write new chunk stream header, fmt is 0
			var pheader *Buffer = r.outHeaderFmt0.Reset()
			pheader.WriteByte(0x00 | byte(msg.PerferCid&0x3F))

			// chunk message header, 11 bytes
			// timestamp, 3bytes, big-endian
			if msg.Header.Timestamp > RTMP_EXTENDED_TIMESTAMP {
				pheader.WriteUInt24(uint32(0xFFFFFF))
			} else {
				pheader.WriteUInt24(uint32(msg.Header.Timestamp))
			}

			// message_length, 3bytes, big-endian
			// message_type, 1bytes
			// message_length, 3bytes, little-endian
			pheader.WriteUInt24(msg.Header.PayloadLength).WriteByte(msg.Header.MessageType).WriteUInt32Le(msg.Header.StreamId)

			// chunk extended timestamp header, 0 or 4 bytes, big-endian
			if msg.Header.Timestamp > RTMP_EXTENDED_TIMESTAMP {
				pheader.WriteUInt32(uint32(msg.Header.Timestamp))
			}

			real_header = r.outHeaderFmt0.WrittenBytes()
		} else {
			// write no message header chunk stream, fmt is 3
			var pheader *Buffer = r.outHeaderFmt3.Reset()
			pheader.WriteByte(0xC0 | byte(msg.PerferCid&0x3F))

			// chunk extended timestamp header, 0 or 4 bytes, big-endian
			// 6.1.3. Extended Timestamp
			// This field is transmitted only when the normal time stamp in the
			// chunk message header is set to 0x00ffffff. If normal time stamp is
			// set to any value less than 0x00ffffff, this field MUST NOT be
			// present. This field MUST NOT be present if the timestamp field is not
			// present. Type 3 chunks MUST NOT have this field.
			// adobe changed for Type3 chunk:
			//		FMLE always sendout the extended-timestamp,
			// 		must send the extended-timestamp to FMS,
			//		must send the extended-timestamp to flash-player.
			// @see: ngx_rtmp_prepare_message
			// @see: http://blog.csdn.net/win_lin/article/details/13363699
			if msg.Header.Timestamp > RTMP_EXTENDED_TIMESTAMP {
				pheader.WriteUInt32(uint32(msg.Header.Timestamp))
			}

			real_header = r.outHeaderFmt3.WrittenBytes()
		}

		// sendout header
		if _, err = r.conn.Write(real_header); err != nil {
			return
		}

		// sendout payload
		if len(msg.Payload) > 0 {
			payload_size := len(msg.Payload) - msg.SentPayloadLength
			payload_size = int(math.Min(float64(r.outChunkSize), float64(payload_size)))

			data := msg.Payload[msg.SentPayloadLength : msg.SentPayloadLength+payload_size]
			if _, err = r.conn.Write(data); err != nil {
				return
			}

			// consume sendout bytes when not empty packet.
			msg.SentPayloadLength += payload_size
		}
	}

	return
}

/**
* recv a message with raw/undecoded payload from peer.
* the payload is not decoded, use srs_rtmp_expect_message<T> if requires
* specifies message.
 */
func (r *protocol) RecvMessage() (msg *Message, err error) {
	var ok bool
	// bug: will hang in here.
	if msg, ok = <-r.msg_in_queue; ok {
		return
	}

	if r.msg_io_err != nil {
		err = r.msg_io_err
		return
	}

	err = Error{code: ERROR_GO_PROTOCOL_DESTROYED, desc: "recv msg from destroyed stack"}
	return
}

/**
* decode the message, return the decoded rtmp packet.
 */
// @see: SrsCommonMessage.decode_packet(SrsProtocol* protocol)
func (r *protocol) DecodeMessage(msg *Message) (pkt interface{}, err error) {
	if msg == nil || msg.Payload == nil {
		return
	}

	pkt, err = DecodePacket(r, msg.Header, msg.Payload)
	return
}

/**
* expect a specified message by v, drop others util got specified one.
 */
func (r *protocol) ExpectPacket(v interface{}) (msg *Message, err error) {
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	if rv.Kind() != reflect.Ptr {
		err = Error{code: ERROR_GO_REFLECT_PTR_REQUIRES, desc: "param must be ptr for expect message"}
		return
	}
	if rv.IsNil() {
		err = Error{code: ERROR_GO_REFLECT_NEVER_NIL, desc: "param should never be nil"}
		return
	}
	if !rv.Elem().CanSet() {
		err = Error{code: ERROR_GO_REFLECT_CAN_SET, desc: "param should be settable"}
		return
	}

	for {
		if msg, err = r.RecvMessage(); err != nil {
			return
		}
		var pkt interface{}
		if pkt, err = r.DecodeMessage(msg); err != nil {
			return
		}
		if pkt == nil {
			continue
		}

		// check the convertible and convert to the value or ptr value.
		// for example, the v like the c++ code: Msg**v
		pkt_rt := reflect.TypeOf(pkt)
		if pkt_rt.ConvertibleTo(rt) {
			// directly match, the pkt is like c++: Msg**pkt
			// set the v by: *v = *pkt
			rv.Elem().Set(reflect.ValueOf(pkt).Elem())
			return
		}
		if pkt_rt.ConvertibleTo(rt.Elem()) {
			// ptr match, the pkt is like c++: Msg*pkt
			// set the v by: *v = pkt
			rv.Elem().Set(reflect.ValueOf(pkt))
			return
		}
	}

	return
}

func (r *protocol) EncodeMessage(pkt Encoder) (cid int, msg *Message, err error) {
	msg = NewMessage()

	cid = pkt.GetPerferCid()

	size := pkt.GetSize()
	if size <= 0 {
		return
	}

	b := make([]byte, size)
	s := NewRtmpStream(b)
	if err = pkt.Encode(s); err != nil {
		return
	}

	msg.Header.MessageType = pkt.GetMessageType()
	msg.Header.PayloadLength = uint32(size)
	msg.Payload = b

	return
}

func (r *protocol) SendPacket(pkt Encoder, stream_id uint32) (err error) {
	var msg *Message = nil

	// if pkt is encoder, encode packet to message.
	var cid int
	if cid, msg, err = r.EncodeMessage(pkt); err != nil {
		return
	}
	msg.PerferCid = cid

	if err = r.SendMessage(msg, stream_id); err != nil {
		return
	}

	if err = r.on_send_message(pkt); err != nil {
		return
	}
	return
}

func (r *protocol) SendMessage(pkt *Message, stream_id uint32) (err error) {
	var msg *Message = pkt

	if msg == nil {
		return Error{code: ERROR_GO_RTMP_NOT_SUPPORT_MSG, desc: "message not support send"}
	}
	if stream_id > 0 {
		msg.Header.StreamId = stream_id
	}

	// let me see the panic.
	// defer func() {
	// 	if re := recover(); re != nil {
	// 		if _, ok := re.(runtime.Error); ok {
	// 			// write to closed channel
	// 			if err == nil {
	// 				err = r.msg_io_err
	// 			}
	// 			return
	// 		}
	// 		panic(re)
	// 	}
	// }()

	r.msg_out_queue <- msg
	return
}

func (r *protocol) on_send_message(pkt Encoder) (err error) {
	if pkt, ok := pkt.(*SetChunkSizePacket); ok {
		r.outChunkSize = pkt.ChunkSize
		return
	}

	if pkt, ok := pkt.(*ConnectAppPacket); ok {
		r.requests[pkt.TransactionId] = pkt.CommandName
		return
	}

	if pkt, ok := pkt.(*CreateStreamPacket); ok {
		r.requests[pkt.TransactionId] = pkt.CommandName
		return
	}
	return
}

func (r *protocol) on_recv_message(msg *Message) (err error) {
	// acknowledgement
	if r.inAckSize.ShouldAckRead(r.conn.RecvBytes()) {
		return r.response_acknowledgement_message()
	}

	// decode the msg if needed
	var pkt interface{}
	if msg.Header.IsSetChunkSize() || msg.Header.IsUserControlMessage() || msg.Header.IsWindowAcknowledgementSize() {
		if pkt, err = r.DecodeMessage(msg); err != nil {
			return
		}
	}

	if pkt, ok := pkt.(*SetChunkSizePacket); ok {
		r.inChunkSize = pkt.ChunkSize
		return
	}

	if pkt, ok := pkt.(*SetWindowAckSizePacket); ok {
		if pkt.AcknowledgementWindowSize > 0 {
			r.inAckSize.ack_window_size = pkt.AcknowledgementWindowSize
		}
		return
	}

	// TODO: FIXME: implements it

	return
}

func (r *protocol) HistoryRequestName(transaction_id float64) (request_name string) {
	request_name, _ = r.requests[transaction_id]
	return
}

func (r *protocol) recv_interlaced_message() (msg *Message, err error) {
	var format byte
	var bh_size, mh_size, cid int

	// chunk stream basic header.
	if format, cid, bh_size, err = r.read_basic_header(); err != nil {
		return
	}

	// get the cached chunk stream.
	chunk, ok := r.chunkStreams[cid]
	if !ok {
		chunk = NewChunkStream(cid)
		r.chunkStreams[cid] = chunk
	}

	// chunk stream message header
	if mh_size, err = r.read_message_header(chunk, format); err != nil {
		return
	}

	// read msg payload from chunk stream.
	if msg, err = r.read_message_payload(chunk, bh_size, mh_size); err != nil {
		return
	}

	// set the perfer cid of message
	if msg != nil {
		msg.PerferCid = cid
	}

	return
}

func (r *protocol) read_basic_header() (format byte, cid int, bh_size int, err error) {
	if err = r.buffer.EnsureBufferBytes(1); err != nil {
		return
	}

	format = r.buffer.ReadByte()
	cid = int(format) & 0x3f
	format = (format >> 6) & 0x03
	bh_size = 1

	if cid == 0 {
		if err = r.buffer.EnsureBufferBytes(1); err != nil {
			return
		}
		cid = 64
		cid += int(r.buffer.ReadByte())
		bh_size = 2
	} else if cid == 1 {
		if err = r.buffer.EnsureBufferBytes(2); err != nil {
			return
		}

		cid = 64
		cid += int(r.buffer.ReadByte())
		cid += int(r.buffer.ReadByte()) * 256
		bh_size = 3
	}

	return
}

func (r *protocol) read_message_header(chunk *ChunkStream, format byte) (mh_size int, err error) {
	/**
	* we should not assert anything about fmt, for the first packet.
	* (when first packet, the chunk->msg is NULL).
	* the fmt maybe 0/1/2/3, the FMLE will send a 0xC4 for some audio packet.
	* the previous packet is:
	* 	04 			// fmt=0, cid=4
	* 	00 00 1a 	// timestamp=26
	*	00 00 9d 	// payload_length=157
	* 	08 			// message_type=8(audio)
	* 	01 00 00 00 // stream_id=1
	* the current packet maybe:
	* 	c4 			// fmt=3, cid=4
	* it's ok, for the packet is audio, and timestamp delta is 26.
	* the current packet must be parsed as:
	* 	fmt=0, cid=4
	* 	timestamp=26+26=52
	* 	payload_length=157
	* 	message_type=8(audio)
	* 	stream_id=1
	* so we must update the timestamp even fmt=3 for first packet.
	 */
	// fresh packet used to update the timestamp even fmt=3 for first packet.
	is_fresh_packet := false
	if chunk.Msg == nil {
		is_fresh_packet = true
	}

	// but, we can ensure that when a chunk stream is fresh,
	// the fmt must be 0, a new stream.
	if chunk.MsgCount == 0 && format != RTMP_FMT_TYPE0 {
		err = Error{code: ERROR_RTMP_CHUNK_START, desc: "protocol error, fmt of first chunk must be 0"}
		return
	}

	// when exists cache msg, means got an partial message,
	// the fmt must not be type0 which means new message.
	if chunk.Msg != nil && format == RTMP_FMT_TYPE0 {
		err = Error{code: ERROR_RTMP_CHUNK_START, desc: "protocol error, unexpect start of new chunk"}
		return
	}

	// create msg when new chunk stream start
	if chunk.Msg == nil {
		chunk.Msg = NewMessage()
	}

	// read message header from socket to buffer.
	mh_sizes := []int{11, 7, 3, 0}
	mh_size = mh_sizes[int(format)]
	if err = r.buffer.EnsureBufferBytes(mh_size); err != nil {
		return
	}

	// parse the message header.
	// see also: ngx_rtmp_recv
	if format <= RTMP_FMT_TYPE2 {
		chunk.Header.TimestampDelta = r.buffer.ReadUInt24()

		// fmt: 0
		// timestamp: 3 bytes
		// If the timestamp is greater than or equal to 16777215
		// (hexadecimal 0x00ffffff), this value MUST be 16777215, and the
		// ‘extended timestamp header’ MUST be present. Otherwise, this value
		// SHOULD be the entire timestamp.
		//
		// fmt: 1 or 2
		// timestamp delta: 3 bytes
		// If the delta is greater than or equal to 16777215 (hexadecimal
		// 0x00ffffff), this value MUST be 16777215, and the ‘extended
		// timestamp header’ MUST be present. Otherwise, this value SHOULD be
		// the entire delta.
		if chunk.ExtendedTimestamp = false; chunk.Header.TimestampDelta >= RTMP_EXTENDED_TIMESTAMP {
			chunk.ExtendedTimestamp = true
		}
		if chunk.ExtendedTimestamp {
			// Extended timestamp: 0 or 4 bytes
			// This field MUST be sent when the normal timsestamp is set to
			// 0xffffff, it MUST NOT be sent if the normal timestamp is set to
			// anything else. So for values less than 0xffffff the normal
			// timestamp field SHOULD be used in which case the extended timestamp
			// MUST NOT be present. For values greater than or equal to 0xffffff
			// the normal timestamp field MUST NOT be used and MUST be set to
			// 0xffffff and the extended timestamp MUST be sent.
			//
			// if extended timestamp, the timestamp must >= RTMP_EXTENDED_TIMESTAMP
			// we set the timestamp to RTMP_EXTENDED_TIMESTAMP to identify we
			// got an extended timestamp.
			chunk.Header.Timestamp = RTMP_EXTENDED_TIMESTAMP
		} else {
			if format == RTMP_FMT_TYPE0 {
				// 6.1.2.1. Type 0
				// For a type-0 chunk, the absolute timestamp of the message is sent
				// here.
				chunk.Header.Timestamp = uint64(chunk.Header.TimestampDelta)
			} else {
				// 6.1.2.2. Type 1
				// 6.1.2.3. Type 2
				// For a type-1 or type-2 chunk, the difference between the previous
				// chunk's timestamp and the current chunk's timestamp is sent here.
				chunk.Header.Timestamp += uint64(chunk.Header.TimestampDelta)
			}
		}

		if format <= RTMP_FMT_TYPE1 {
			chunk.Header.PayloadLength = r.buffer.ReadUInt24()

			// if msg exists in cache, the size must not changed.
			if chunk.Msg.Payload != nil && len(chunk.Msg.Payload) != int(chunk.Header.PayloadLength) {
				err = Error{code: ERROR_RTMP_PACKET_SIZE, desc: "cached message size should never change"}
				return
			}

			chunk.Header.MessageType = r.buffer.ReadByte()

			if format == RTMP_FMT_TYPE0 {
				chunk.Header.StreamId = r.buffer.ReadUInt32Le()
			}
		}
	} else {
		// update the timestamp even fmt=3 for first stream
		if is_fresh_packet && !chunk.ExtendedTimestamp {
			chunk.Header.Timestamp += uint64(chunk.Header.TimestampDelta)
		}
	}

	if chunk.ExtendedTimestamp {
		mh_size += 4
		if err = r.buffer.EnsureBufferBytes(4); err != nil {
			return
		}

		// ffmpeg/librtmp may donot send this filed, need to detect the value.
		// @see also: http://blog.csdn.net/win_lin/article/details/13363699
		timestamp := r.buffer.ReadUInt32()

		// compare to the chunk timestamp, which is set by chunk message header
		// type 0,1 or 2.
		if chunk.Header.Timestamp > RTMP_EXTENDED_TIMESTAMP && chunk.Header.Timestamp != uint64(timestamp) {
			mh_size -= 4
			r.buffer.Skip(-4)
		} else {
			chunk.Header.Timestamp = uint64(timestamp)
		}
	}

	// valid message
	if int32(chunk.Header.PayloadLength) < 0 {
		err = Error{code: ERROR_RTMP_MSG_INVLIAD_SIZE, desc: "chunk packet should never be negative"}
		return
	}

	// copy header to msg
	copy := *chunk.Header
	chunk.Msg.Header = &copy

	// increase the msg count, the chunk stream can accept fmt=1/2/3 message now.
	chunk.MsgCount++

	return
}

func (r *protocol) read_message_payload(chunk *ChunkStream, bh_size int, mh_size int) (msg *Message, err error) {
	// empty message
	if int32(chunk.Header.PayloadLength) <= 0 {
		msg = chunk.Msg
		chunk.Msg = nil
		err = r.buffer.Consume(mh_size + bh_size)
		return
	}

	// the chunk payload size.
	payload_size := int(chunk.Header.PayloadLength) - chunk.Msg.ReceivedPayloadLength
	payload_size = int(math.Min(float64(payload_size), float64(r.inChunkSize)))

	// create msg payload if not initialized
	if chunk.Msg.Payload == nil {
		chunk.Msg.Payload = make([]byte, chunk.Msg.Header.PayloadLength)
	}

	// read payload to buffer
	if err = r.buffer.EnsureBufferBytes(payload_size); err != nil {
		return
	}
	copy(chunk.Msg.Payload[chunk.Msg.ReceivedPayloadLength:chunk.Msg.ReceivedPayloadLength+payload_size], r.buffer.Read(payload_size))
	chunk.Msg.ReceivedPayloadLength += payload_size
	if err = r.buffer.Consume(mh_size + bh_size + payload_size); err != nil {
		return
	}

	// got entire RTMP message?
	if chunk.Msg.ReceivedPayloadLength == len(chunk.Msg.Payload) {
		msg = chunk.Msg
		chunk.Msg = nil
		return
	}

	return
}

func (r *protocol) response_acknowledgement_message() (err error) {
	// TODO: FIXME: implements it
	return
}

func (r *MessageHeader) IsAmf0Command() bool {
	return r.MessageType == RTMP_MSG_AMF0CommandMessage
}
func (r *MessageHeader) IsAmf3Command() bool {
	return r.MessageType == RTMP_MSG_AMF3CommandMessage
}
func (r *MessageHeader) IsAmf0Data() bool {
	return r.MessageType == RTMP_MSG_AMF0DataMessage
}
func (r *MessageHeader) IsAmf3Data() bool {
	return r.MessageType == RTMP_MSG_AMF3DataMessage
}
func (r *MessageHeader) IsWindowAcknowledgementSize() bool {
	return r.MessageType == RTMP_MSG_WindowAcknowledgementSize
}
func (r *MessageHeader) IsSetChunkSize() bool {
	return r.MessageType == RTMP_MSG_SetChunkSize
}
func (r *MessageHeader) IsUserControlMessage() bool {
	return r.MessageType == RTMP_MSG_UserControlMessage
}
func (r *MessageHeader) IsVideo() bool {
	return r.MessageType == RTMP_MSG_VideoMessage
}
func (r *MessageHeader) IsAudio() bool {
	return r.MessageType == RTMP_MSG_AudioMessage
}
func (r *MessageHeader) IsAggregate() bool {
	return r.MessageType == RTMP_MSG_AggregateMessage
}

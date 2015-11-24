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
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

/**
* the rtmp message, encode/decode to/from the rtmp stream,
* which contains a message header and a bytes payload.
* the header is MessageHeader, where the payload canbe decoded by RtmpPacket.
 */
// @see: ISrsMessage, SrsCommonMessage, SrsSharedPtrMessage
type Message struct {
	// 4.1. Message Header
	Header *MessageHeader
	// 4.2. Message Payload
	/**
	* The other part which is the payload is the actual data that is
	* contained in the message. For example, it could be some audio samples
	* or compressed video data. The payload format and interpretation are
	* beyond the scope of this document.
	 */
	Payload []byte
	/**
	* the payload is received from connection,
	* when len(Payload) == ReceivedPayloadLength, message receive completed.
	 */
	ReceivedPayloadLength int
	/**
	* get the perfered cid(chunk stream id) which sendout over.
	* set at decoding, and canbe used for directly send message,
	* for example, dispatch to all connections.
	* @see: SrsSharedPtrMessage.SrsSharedPtr.perfer_cid
	 */
	PerferCid int
	/**
	* the payload sent length.
	 */
	SentPayloadLength int
}

func NewMessage() *Message {
	r := &Message{}
	r.Header = &MessageHeader{}
	return r
}

// copy the message, deep copy header and field, share copy the payload
func (r *Message) Copy() *Message {
	copy := &Message{}
	copy_header := *r.Header
	copy.Header = &copy_header
	copy.Payload = r.Payload
	copy.ReceivedPayloadLength = r.ReceivedPayloadLength
	copy.PerferCid = r.PerferCid
	copy.SentPayloadLength = r.SentPayloadLength
	return copy
}

/**
* incoming chunk stream maybe interlaced,
* use the chunk stream to cache the input RTMP chunk streams.
 */
type ChunkStream struct {
	/**
	* represents the basic header fmt,
	* which used to identify the variant message header type.
	 */
	FMT byte
	/**
	* represents the basic header cid,
	* which is the chunk stream id.
	 */
	CId int
	/**
	* cached message header
	 */
	Header *MessageHeader
	/**
	* whether the chunk message header has extended timestamp.
	 */
	ExtendedTimestamp bool
	/**
	* partially read message.
	 */
	Msg *Message
	/**
	* decoded msg count, to identify whether the chunk stream is fresh.
	 */
	MsgCount int64
}

func NewChunkStream(cid int) (r *ChunkStream) {
	r = &ChunkStream{}

	r.CId = cid
	r.Header = &MessageHeader{}

	return
}

/**
* the message header for Message,
* the header can be used in chunk stream cache, for the chunk stream header.
* @see: RTMP 4.1. Message Header
 */
type MessageHeader struct {
	/**
	* One byte field to represent the message type. A range of type IDs
	* (1-7) are reserved for protocol control messages.
	 */
	MessageType byte
	/**
	* Three-byte field that represents the size of the payload in bytes.
	* It is set in big-endian format.
	 */
	PayloadLength uint32
	/**
	* Three-byte field that contains a timestamp delta of the message.
	* The 3 bytes are packed in the big-endian order.
	* @remark, only used for decoding message from chunk stream.
	 */
	TimestampDelta uint32
	/**
	* Four-byte field that identifies the stream of the message. These
	* bytes are set in little-endian format.
	 */
	StreamId uint32

	/**
	* Four-byte field that contains a timestamp of the message.
	* The 4 bytes are packed in the big-endian order.
	* @remark, used as calc timestamp when decode and encode time.
	* @remark, we use 64bits for large time for jitter detect and hls.
	 */
	Timestamp uint64
}

type Protocol interface {
	/**
	* destroy the protocol stack, close channels, stop goroutines.
	 */
	Destroy()
	/**
	* get the message input channel,
	* the input goroutine decode and put message into the input channel,
	* where user can select the channel to recv message.
	 */
	MessageInputChannel() chan *Message
	/**
	* do simple handshake with client, user can try simple/complex interlace,
	* that is, try complex handshake first, use simple if complex handshake failed.
	* when handshake success, start the message input/outout goroutines
	 */
	SimpleHandshake2Client() (err error)
	/**
	* recv message from connection.
	* the payload of message is []byte, user can decode it by DecodeMessage.
	 */
	RecvMessage() (msg *Message, err error)
	/**
	* decode the received message to pkt.
	 */
	DecodeMessage(msg *Message) (pkt interface{}, err error)
	/**
	* expect specified packet by v, where v must be a ptr,
	* protocol stack will RecvMessage from connection and DecodeMessage(msg) to pkt,
	* then convert/set msg to v if type matched, or drop the message and try again.
	* for example:
	* 		var pkt *ConnectAppPacket
	*		_, err = r.protocol.ExpectPacket(&pkt)
	* 		// use the decoded pkt contains the connect app info.
	 */
	ExpectPacket(v interface{}) (msg *Message, err error)
	/**
	* encode the packet to message, then send out by SendMessage.
	* return the cid which packet prefer.
	 */
	//EncodeMessage(pkt Encoder) (cid int, msg *Message, err error)
	/**
	* send message to peer over rtmp connection.
	* if pkt is Encoder, encode the pkt to Message and send out.
	* if pkt is Message already, directly send it out.
	 */
	SendPacket(pkt Encoder, stream_id uint32) (err error)
	SendMessage(pkt *Message, stream_id uint32) (err error)
}

/**
* max rtmp header size:
* 	1bytes basic header,
* 	11bytes message header,
* 	4bytes timestamp header,
* that is, 1+11+4=16bytes.
 */
const RTMP_MAX_FMT0_HEADER_SIZE = 16

/**
* max rtmp header size:
* 	1bytes basic header,
* 	4bytes timestamp header,
* that is, 1+4=5bytes.
 */
const RTMP_MAX_FMT3_HEADER_SIZE = 5

// the buffer size of msg channel
const RTMP_MSG_CHANNEL_BUFFER = 100

/**
* create the rtmp protocol.
 */
func NewProtocol(conn *net.TCPConn) (Protocol, error) {
	r := &protocol{}

	r.conn = NewSocket(conn)
	r.chunkStreams = map[int]*ChunkStream{}
	r.buffer = NewRtmpBuffer(r.conn)
	r.handshake = &Handshake{}

	r.inChunkSize = RTMP_DEFAULT_CHUNK_SIZE
	r.outChunkSize = r.inChunkSize
	r.outHeaderFmt0 = NewRtmpStream(make([]byte, RTMP_MAX_FMT0_HEADER_SIZE))
	r.outHeaderFmt3 = NewRtmpStream(make([]byte, RTMP_MAX_FMT3_HEADER_SIZE))

	// r.msg_in_lock = &sync.Mutex{}
	// r.msg_out_lock = &sync.Mutex{}
	r.closeChan = make(chan struct{})
	r.msg_in_queue = make(chan *Message, RTMP_MSG_CHANNEL_BUFFER)
	r.msg_out_queue = make(chan *Message, RTMP_MSG_CHANNEL_BUFFER)

	rand.Seed(time.Now().UnixNano())

	return r, nil
}

/**
* the payload codec by the RtmpPacket.
* @see: RTMP 4.2. Message Payload
 */
// @see: SrsPacket
/**
* the decoded message payload.
* @remark we seperate the packet from message,
*		for the packet focus on logic and domain data,
*		the message bind to the protocol and focus on protocol, such as header.
* 		we can merge the message and packet, using OOAD hierachy, packet extends from message,
* 		it's better for me to use components -- the message use the packet as payload.
 */
type Decoder interface {
	/**
	* decode the packet from the s, which is created by rtmp message.
	 */
	Decode(s *Buffer) (err error)
}

/**
* encode the rtmp packet to payload of rtmp message.
 */
type Encoder interface {
	/**
	* get the rtmp chunk cid the packet perfered.
	 */
	GetPerferCid() (v int)
	/**
	* get packet message type
	 */
	GetMessageType() (v byte)
	/**
	* get the size of packet, to create the *HPBuffer.
	 */
	GetSize() (v int)
	/**
	* encode the packet to s, which is created by size=GetSize()
	 */
	Encode(s *Buffer) (err error)
}

func DecodePacket(r *protocol, header *MessageHeader, payload []byte) (packet interface{}, err error) {
	var pkt Decoder = nil
	var stream *Buffer = NewRtmpStream(payload)

	// decode specified packet type
	if header.IsAmf0Command() || header.IsAmf3Command() || header.IsAmf0Data() || header.IsAmf3Data() {
		// skip 1bytes to decode the amf3 command.
		if header.IsAmf3Command() && stream.Requires(1) {
			stream = NewRtmpStream(payload[1:])
		}

		amf0_codec := NewAmf0Codec(stream)

		// amf0 command message.
		// need to read the command name.
		var command string
		if command, err = amf0_codec.ReadString(); err != nil {
			return
		}

		// result/error packet
		if command == AMF0_COMMAND_RESULT || command == AMF0_COMMAND_ERROR {
			var transaction_id float64
			if transaction_id, err = amf0_codec.ReadNumber(); err != nil {
				return
			}

			// reset to zero to restart decode.
			stream.Reset()

			var request_name string
			if request_name = r.HistoryRequestName(transaction_id); request_name == "" {
				err = Error{code: ERROR_RTMP_NO_REQUEST, desc: "decode AMF0/AMF3 transaction request failed"}
				return
			}

			// TODO: FIXME: implements it
		}

		// reset to zero to restart decode.
		stream.Reset()

		// decode command object.
		switch command {
		case AMF0_COMMAND_CONNECT:
			pkt = NewConnectAppPacket()
		case AMF0_COMMAND_CREATE_STREAM:
			pkt = NewCreateStreamPacket()
		case AMF0_COMMAND_PLAY:
			pkt = NewPlayPacket()
		case AMF0_COMMAND_PUBLISH:
			pkt = NewPublishPacket()
		case AMF0_COMMAND_CLOSE_STREAM:
			pkt = NewCloseStreamPacket()
		case AMF0_COMMAND_RELEASE_STREAM:
			pkt = NewFMLEStartPacket()
		case AMF0_COMMAND_FC_PUBLISH:
			pkt = NewFMLEStartPacket()
		case AMF0_COMMAND_UNPUBLISH:
			pkt = NewFMLEStartPacket()
		}
		// TODO: FIXME: implements it
	} else if header.IsWindowAcknowledgementSize() {
		pkt = NewSetWindowAckSizePacket()
	} else if header.IsUserControlMessage() {
		pkt = NewUserControlPacket()
	} else if header.IsSetChunkSize() {
		pkt = NewSetChunkSizePacket()
	}
	// TODO: FIXME: implements it

	if err == nil && pkt != nil {
		packet, err = pkt, pkt.Decode(stream)
	}

	return
}

/**
* 4.1.1. connect
* The client sends the connect command to the server to request
* connection to a server application instance.
 */
// @see: SrsConnectAppPacket
type ConnectAppPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Object
}

func NewConnectAppPacket() *ConnectAppPacket {
	r := &ConnectAppPacket{}
	r.TransactionId = float64(1.0)
	r.CommandObject = NewAmf0Object()
	return r
}
func (r *ConnectAppPacket) Set(k string, v interface{}) *ConnectAppPacket {
	// if empty or empty object, any value must has content.
	if a := NewAmf0(v); a != nil && a.Size() > 0 {
		r.CommandObject.Set(k, a)
	}
	return r
}

// Decoder
func (r *ConnectAppPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName != AMF0_COMMAND_CONNECT {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=%v, actual=%v", AMF0_COMMAND_CONNECT, r.CommandName)}
	}

	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if r.TransactionId != 1.0 {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: "amf0 decode connect transaction_id failed."}
	}

	if r.CommandObject, err = codec.ReadObject(); err != nil {
		return
	}
	if r.CommandObject == nil {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: "amf0 decode connect command_object failed."}
	}

	return
}

// Encoder
func (r *ConnectAppPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverConnection
}
func (r *ConnectAppPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *ConnectAppPacket) GetSize() (v int) {
	v = Amf0SizeString(r.CommandName)
	v += Amf0SizeNumber()
	v += r.CommandObject.Size()
	return
}
func (r *ConnectAppPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if r.CommandObject.Size() > 0 {
		if err = codec.WriteObject(r.CommandObject); err != nil {
			return
		}
	}
	return
}

/**
* response for SrsConnectAppPacket.
 */
// @see: SrsConnectAppResPacket
type ConnectAppResPacket struct {
	CommandName   string
	TransactionId float64
	Props         *Amf0Object
	Info          *Amf0Object
}

func NewConnectAppResPacket() *ConnectAppResPacket {
	r := &ConnectAppResPacket{}
	r.CommandName = AMF0_COMMAND_RESULT
	r.TransactionId = float64(1.0)
	r.Props = NewAmf0Object()
	r.Info = NewAmf0Object()
	return r
}
func (r *ConnectAppResPacket) PropsSet(k string, v interface{}) *ConnectAppResPacket {
	// if empty or empty object, any value must has content.
	if a := NewAmf0(v); a != nil && a.Size() > 0 {
		r.Props.Set(k, a)
	}
	return r
}
func (r *ConnectAppResPacket) InfoSet(k string, v interface{}) *ConnectAppResPacket {
	// if empty or empty object, any value must has content.
	if a := NewAmf0(v); a != nil && a.Size() > 0 {
		r.Info.Set(k, a)
	}
	return r
}

// Encoder
func (r *ConnectAppResPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverConnection
}
func (r *ConnectAppResPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *ConnectAppResPacket) GetSize() (v int) {
	v = Amf0SizeString(r.CommandName)
	v += Amf0SizeNumber()
	v += r.Props.Size()
	v += r.Info.Size()
	return
}
func (r *ConnectAppResPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if r.Props.Size() > 0 {
		if err = codec.WriteObject(r.Props); err != nil {
			return
		}
	}
	if r.Info.Size() > 0 {
		if err = codec.WriteObject(r.Info); err != nil {
			return
		}
	}
	return
}

/**
* 5.5. Window Acknowledgement Size (5)
* The client or the server sends this message to inform the peer which
* window size to use when sending acknowledgment.
 */
// @see: SrsSetWindowAckSizePacket
type SetWindowAckSizePacket struct {
	AcknowledgementWindowSize uint32
}

func NewSetWindowAckSizePacket() *SetWindowAckSizePacket {
	return &SetWindowAckSizePacket{}
}

// Decoder
func (r *SetWindowAckSizePacket) Decode(s *Buffer) (err error) {
	if !s.Requires(4) {
		err = Error{code: ERROR_RTMP_MESSAGE_DECODE, desc: "decode ack window size failed."}
		return
	}
	r.AcknowledgementWindowSize = s.ReadUInt32()
	return
}

// Encoder
func (r *SetWindowAckSizePacket) GetPerferCid() (v int) {
	return RTMP_CID_ProtocolControl
}
func (r *SetWindowAckSizePacket) GetMessageType() (v byte) {
	return RTMP_MSG_WindowAcknowledgementSize
}
func (r *SetWindowAckSizePacket) GetSize() (v int) {
	return 4
}
func (r *SetWindowAckSizePacket) Encode(s *Buffer) (err error) {
	if !s.Requires(4) {
		return Error{code: ERROR_RTMP_MESSAGE_ENCODE, desc: "encode ack size packet failed."}
	}
	s.WriteUInt32(r.AcknowledgementWindowSize)
	return
}

/**
* 7.1. Set Chunk Size
* Protocol control message 1, Set Chunk Size, is used to notify the
* peer about the new maximum chunk size.
 */
// @see: SrsSetChunkSizePacket
type SetChunkSizePacket struct {
	ChunkSize uint32
}

func NewSetChunkSizePacket() *SetChunkSizePacket {
	r := &SetChunkSizePacket{}
	r.ChunkSize = RTMP_DEFAULT_CHUNK_SIZE
	return r
}

// Decoder
func (r *SetChunkSizePacket) Decode(s *Buffer) (err error) {
	if !s.Requires(4) {
		err = Error{code: ERROR_RTMP_MESSAGE_DECODE, desc: "decode chunk size failed."}
		return
	}
	r.ChunkSize = s.ReadUInt32()

	if r.ChunkSize < RTMP_MIN_CHUNK_SIZE {
		err = Error{code: ERROR_RTMP_CHUNK_SIZE, desc: "atleast min chunk size."}
	}
	if r.ChunkSize > RTMP_MAX_CHUNK_SIZE {
		err = Error{code: ERROR_RTMP_CHUNK_SIZE, desc: "exceed max chunk size."}
	}
	return
}

// Encoder
func (r *SetChunkSizePacket) GetPerferCid() (v int) {
	return RTMP_CID_ProtocolControl
}
func (r *SetChunkSizePacket) GetMessageType() (v byte) {
	return RTMP_MSG_SetChunkSize
}
func (r *SetChunkSizePacket) GetSize() (v int) {
	return 4
}
func (r *SetChunkSizePacket) Encode(s *Buffer) (err error) {
	if !s.Requires(4) {
		return Error{code: ERROR_RTMP_MESSAGE_ENCODE, desc: "encode chunk packet failed."}
	}
	s.WriteUInt32(r.ChunkSize)
	return
}

/**
* 5.6. Set Peer Bandwidth (6)
* The client or the server sends this message to update the output
* bandwidth of the peer.
 */
// @see: SrsSetPeerBandwidthPacket
type SetPeerBandwidthPacket struct {
	Bandwidth     uint32
	BandwidthType byte
}

// Encoder
func (r *SetPeerBandwidthPacket) GetPerferCid() (v int) {
	return RTMP_CID_ProtocolControl
}
func (r *SetPeerBandwidthPacket) GetMessageType() (v byte) {
	return RTMP_MSG_SetPeerBandwidth
}
func (r *SetPeerBandwidthPacket) GetSize() (v int) {
	return 5
}
func (r *SetPeerBandwidthPacket) Encode(s *Buffer) (err error) {
	if !s.Requires(5) {
		return Error{code: ERROR_RTMP_MESSAGE_ENCODE, desc: "encode set bandwidth packet failed."}
	}
	s.WriteUInt32(r.Bandwidth).WriteByte(r.BandwidthType)
	return
}

/**
* 5.6. Set Peer Bandwidth (6)
* The client or the server sends this message to update the output
* bandwidth of the peer.
 */
// @see: SrsOnBWDonePacket
type OnBWDonePacket struct {
	CommandName   string
	TransactionId float64
	Args          *Amf0Any // Null
}

func NewOnBWDonePacket() *OnBWDonePacket {
	r := &OnBWDonePacket{}
	r.CommandName = AMF0_COMMAND_ON_BW_DONE
	r.Args = NewAmf0Null()
	return r
}

// Encoder
func (r *OnBWDonePacket) GetPerferCid() (v int) {
	return RTMP_CID_OverConnection
}
func (r *OnBWDonePacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *OnBWDonePacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + Amf0SizeNumber() + Amf0SizeNullOrUndefined()
}
func (r *OnBWDonePacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)
	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.Args.Write(codec); err != nil {
		return
	}
	return
}

/**
* 4.1.3. createStream
* The client sends this command to the server to create a logical
* channel for message communication The publishing of audio, video, and
* metadata is carried out over stream channel created using the
* createStream command.
 */
// @see: SrsCreateStreamPacket
type CreateStreamPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
}

func NewCreateStreamPacket() *CreateStreamPacket {
	r := &CreateStreamPacket{}
	r.CommandName = AMF0_COMMAND_CREATE_STREAM
	r.TransactionId = 2.0
	r.CommandObject = NewAmf0Null()
	return r
}

// Decoder
func (r *CreateStreamPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName == "" || r.CommandName != AMF0_COMMAND_CREATE_STREAM {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=%v, actual=%v", AMF0_COMMAND_CREATE_STREAM, r.CommandName)}
	}
	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if err = r.CommandObject.Read(codec); err != nil {
		return
	}
	return
}

// Encoder
func (r *CreateStreamPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverConnection
}
func (r *CreateStreamPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *CreateStreamPacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + Amf0SizeNumber() + Amf0SizeNullOrUndefined()
}
func (r *CreateStreamPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.CommandObject.Write(codec); err != nil {
		return
	}
	return
}

/**
* response for SrsCreateStreamPacket.
 */
// @see: SrsCreateStreamResPacket
type CreateStreamResPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
	StreamId      float64
}

func NewCreateStreamResPacket(transaction_id float64, stream_id float64) *CreateStreamResPacket {
	r := &CreateStreamResPacket{}
	r.CommandName = AMF0_COMMAND_RESULT
	r.TransactionId = transaction_id
	r.CommandObject = NewAmf0Null()
	r.StreamId = stream_id
	return r
}

// Decoder
func (r *CreateStreamResPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName == "" || r.CommandName != AMF0_COMMAND_RESULT {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=%v, actual=%v", AMF0_COMMAND_RESULT, r.CommandName)}
	}
	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if err = r.CommandObject.Read(codec); err != nil {
		return
	}
	if r.StreamId, err = codec.ReadNumber(); err != nil {
		return
	}
	return
}

// Encoder
func (r *CreateStreamResPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverConnection
}
func (r *CreateStreamResPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *CreateStreamResPacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + Amf0SizeNumber() + Amf0SizeNullOrUndefined() + Amf0SizeNumber()
}
func (r *CreateStreamResPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.CommandObject.Write(codec); err != nil {
		return
	}
	if err = codec.WriteNumber(r.StreamId); err != nil {
		return
	}
	return
}

/**
* 4.2.1. play
* The client sends this command to the server to play a stream.
 */
// @see: SrsPlayPacket
type PlayPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
	StreamName    string
	Start         float64
	Duration      float64
	Reset         bool
}

func NewPlayPacket() *PlayPacket {
	r := &PlayPacket{}
	r.CommandName = AMF0_COMMAND_PLAY
	r.CommandObject = NewAmf0Null()
	r.Start = -2
	r.Duration = -1
	r.Reset = true
	return r
}

// Decoder
func (r *PlayPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName == "" || r.CommandName != AMF0_COMMAND_PLAY {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=%v, actual=%v", AMF0_COMMAND_PLAY, r.CommandName)}
	}
	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if err = r.CommandObject.Read(codec); err != nil {
		return
	}
	if r.StreamName, err = codec.ReadString(); err != nil {
		return
	}
	if !s.Empty() {
		if r.Start, err = codec.ReadNumber(); err != nil {
			return
		}
	}
	if !s.Empty() {
		if r.Duration, err = codec.ReadNumber(); err != nil {
			return
		}
	}

	if s.Empty() {
		return
	}
	var reset_value = Amf0Any{}
	if err = reset_value.Read(codec); err != nil {
		return
	}
	if v, ok := reset_value.Boolean(); ok {
		r.Reset = v
	} else if v, ok := reset_value.Number(); ok {
		r.Reset = (v != 0)
	} else {
		err = Error{code: ERROR_RTMP_AMF0_DECODE, desc: "amf0 invalid type, requires number or bool"}
	}
	return
}

// Encoder
func (r *PlayPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverStream
}
func (r *PlayPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *PlayPacket) GetSize() (v int) {
	v = Amf0SizeString(r.CommandName) + Amf0SizeNumber() + Amf0SizeNullOrUndefined() + Amf0SizeString(r.StreamName)
	v += Amf0SizeNumber() + Amf0SizeNumber() + Amf0SizeBoolean()
	return
}
func (r *PlayPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.CommandObject.Write(codec); err != nil {
		return
	}
	if err = codec.WriteString(r.StreamName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.Start); err != nil {
		return
	}
	if err = codec.WriteNumber(r.Duration); err != nil {
		return
	}
	if err = codec.WriteBoolean(r.Reset); err != nil {
		return
	}
	return
}

/**
* FMLE/flash publish
* 4.2.6. Publish
* The client sends the publish command to publish a named stream to the
* server. Using this name, any client can play this stream and receive
* the published audio, video, and data messages.
 */
// @see: SrsPublishPacket
type PublishPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
	StreamName    string
	// optional, default to live.
	StreamType string
}

func NewPublishPacket() *PublishPacket {
	r := &PublishPacket{}
	r.CommandName = AMF0_COMMAND_PUBLISH
	r.CommandObject = NewAmf0Null()
	r.StreamType = "live"
	return r
}

// Decoder
func (r *PublishPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName == "" || r.CommandName != AMF0_COMMAND_PUBLISH {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=%v, actual=%v", AMF0_COMMAND_PUBLISH, r.CommandName)}
	}
	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if err = r.CommandObject.Read(codec); err != nil {
		return
	}
	if r.StreamName, err = codec.ReadString(); err != nil {
		return
	}
	if !s.Empty() {
		if r.StreamType, err = codec.ReadString(); err != nil {
			return
		}
	}
	return
}

// Encoder
func (r *PublishPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverStream
}
func (r *PublishPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *PublishPacket) GetSize() (v int) {
	v = Amf0SizeString(r.CommandName) + Amf0SizeNumber() + Amf0SizeNullOrUndefined() + Amf0SizeString(r.StreamName)
	v += Amf0SizeString(r.StreamType)
	return
}
func (r *PublishPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.CommandObject.Write(codec); err != nil {
		return
	}
	if err = codec.WriteString(r.StreamName); err != nil {
		return
	}
	if err = codec.WriteString(r.StreamType); err != nil {
		return
	}
	return
}

// 3.7. User Control message
// @see: SrcPCUCEventType
const (
	// generally, 4bytes event-data
	PCUCStreamBegin      = 0
	PCUCStreamEOF        = 1
	PCUCStreamDry        = 2
	PCUCSetBufferLength  = 3 // 8bytes event-data
	PCUCStreamIsRecorded = 4
	PCUCPingRequest      = 6
	PCUCPingResponse     = 7
)

/**
* for the EventData is 4bytes.
* Stream Begin(=0)			4-bytes stream ID
* Stream EOF(=1)			4-bytes stream ID
* StreamDry(=2)				4-bytes stream ID
* SetBufferLength(=3)		8-bytes 4bytes stream ID, 4bytes buffer length.
* StreamIsRecorded(=4)		4-bytes stream ID
* PingRequest(=6)			4-bytes timestamp local server time
* PingResponse(=7)			4-bytes timestamp received ping request.
*
* 3.7. User Control message
* +------------------------------+-------------------------
* | Event Type ( 2- bytes ) | Event Data
* +------------------------------+-------------------------
* Figure 5 Pay load for the ‘User Control Message’.
 */
// @see: SrsUserControlPacket
type UserControlPacket struct {
	// @see: SrcPCUCEventType
	// for example, PCUCStreamBegin
	EventType uint16
	EventData uint32
	/**
	* 4bytes if event_type is SetBufferLength; otherwise 0.
	 */
	ExtraData uint32
}

func NewUserControlPacket() *UserControlPacket {
	r := &UserControlPacket{}
	return r
}

// Decoder
func (r *UserControlPacket) Decode(s *Buffer) (err error) {
	if !s.Requires(6) {
		return Error{code: ERROR_RTMP_MESSAGE_DECODE, desc: "decode user control failed"}
	}

	r.EventType = s.ReadUInt16()
	r.EventData = s.ReadUInt32()

	if r.EventType != PCUCSetBufferLength {
		return
	}

	if !s.Requires(4) {
		return Error{code: ERROR_RTMP_MESSAGE_DECODE, desc: "decode PCUC set buffer length failed"}
	}
	r.ExtraData = s.ReadUInt32()
	return
}

// Encoder
func (r *UserControlPacket) GetPerferCid() (v int) {
	return RTMP_CID_ProtocolControl
}
func (r *UserControlPacket) GetMessageType() (v byte) {
	return RTMP_MSG_UserControlMessage
}
func (r *UserControlPacket) GetSize() (v int) {
	if r.EventType == PCUCSetBufferLength {
		return 2 + 4 + 4
	} else {
		return 2 + 4
	}
}
func (r *UserControlPacket) Encode(s *Buffer) (err error) {
	if !s.Requires(6) {
		return Error{code: ERROR_RTMP_MESSAGE_ENCODE, desc: "encode user control failed"}
	}
	s.WriteUInt16(r.EventType).WriteUInt32(r.EventData)

	// when event type is set buffer length,
	// write the extra buffer length.
	if r.EventType != PCUCSetBufferLength {
		return
	}

	if !s.Requires(4) {
		return Error{code: ERROR_RTMP_MESSAGE_ENCODE, desc: "encode PCUC set buffer length failed"}
	}
	s.WriteUInt32(r.ExtraData)
	return
}

/**
* onStatus command, AMF0 Call
* @remark, user must set the stream_id by SrsMessage.set_packet().
 */
// @see: SrsOnStatusCallPacket
type OnStatusCallPacket struct {
	CommandName   string
	TransactionId float64
	Args          *Amf0Any // Null
	Data          *Amf0Object
}

func NewOnStatusCallPacket() *OnStatusCallPacket {
	r := &OnStatusCallPacket{}
	r.CommandName = AMF0_COMMAND_ON_STATUS
	r.Args = NewAmf0Null()
	r.Data = NewAmf0Object()
	return r
}
func (r *OnStatusCallPacket) Set(k string, v interface{}) *OnStatusCallPacket {
	// if empty or empty object, any value must has content.
	if a := NewAmf0(v); a != nil && a.Size() > 0 {
		r.Data.Set(k, a)
	}
	return r
}

// Encoder
func (r *OnStatusCallPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverStream
}
func (r *OnStatusCallPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *OnStatusCallPacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + Amf0SizeNumber() + Amf0SizeNullOrUndefined() + r.Data.Size()
}
func (r *OnStatusCallPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.Args.Write(codec); err != nil {
		return
	}
	if err = r.Data.Write(codec); err != nil {
		return
	}
	return
}

/**
* AMF0Data RtmpSampleAccess
* @remark, user must set the stream_id by SrsMessage.set_packet().
 */
// @see: SrsSampleAccessPacket
type SampleAccessPacket struct {
	CommandName       string
	VideoSampleAccess bool
	AudioSampleAccess bool
}

func NewSampleAccessPacket() *SampleAccessPacket {
	r := &SampleAccessPacket{}
	r.CommandName = AMF0_DATA_SAMPLE_ACCESS
	return r
}

// Encoder
func (r *SampleAccessPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverStream
}
func (r *SampleAccessPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0DataMessage
}
func (r *SampleAccessPacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + Amf0SizeBoolean() + Amf0SizeBoolean()
}
func (r *SampleAccessPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteBoolean(r.VideoSampleAccess); err != nil {
		return
	}
	if err = codec.WriteBoolean(r.AudioSampleAccess); err != nil {
		return
	}
	return
}

/**
* onStatus data, AMF0 Data
* @remark, user must set the stream_id by SrsMessage.set_packet().
 */
// @see: SrsOnStatusDataPacket
type OnStatusDataPacket struct {
	CommandName string
	Data        *Amf0Object
}

func NewOnStatusDataPacket() *OnStatusDataPacket {
	r := &OnStatusDataPacket{}
	r.CommandName = AMF0_COMMAND_ON_STATUS
	r.Data = NewAmf0Object()
	return r
}
func (r *OnStatusDataPacket) Set(k string, v interface{}) *OnStatusDataPacket {
	// if empty or empty object, any value must has content.
	if a := NewAmf0(v); a != nil && a.Size() > 0 {
		r.Data.Set(k, a)
	}
	return r
}

// Encoder
func (r *OnStatusDataPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverStream
}
func (r *OnStatusDataPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0DataMessage
}
func (r *OnStatusDataPacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + r.Data.Size()
}
func (r *OnStatusDataPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = r.Data.Write(codec); err != nil {
		return
	}
	return
}

/**
* client close stream packet.
 */
// @see: SrsCloseStreamPacket
type CloseStreamPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
}

func NewCloseStreamPacket() *CloseStreamPacket {
	r := &CloseStreamPacket{}
	r.CommandName = AMF0_COMMAND_CLOSE_STREAM
	r.CommandObject = NewAmf0Null()
	return r
}

// Decoder
func (r *CloseStreamPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName != AMF0_COMMAND_CLOSE_STREAM {
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=%v, actual=%v", AMF0_COMMAND_CLOSE_STREAM, r.CommandName)}
	}

	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if err = r.CommandObject.Read(codec); err != nil {
		return
	}

	return
}

/**
* FMLE start publish: ReleaseStream/PublishStream
 */
// @see: SrsFMLEStartPacket
type FMLEStartPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
	StreamName    string
}

func NewFMLEStartPacket() *FMLEStartPacket {
	r := &FMLEStartPacket{}
	r.CommandName = AMF0_COMMAND_RELEASE_STREAM
	r.CommandObject = NewAmf0Null()
	return r
}

// Decoder
func (r *FMLEStartPacket) Decode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if r.CommandName, err = codec.ReadString(); err != nil {
		return
	}
	if r.CommandName != AMF0_COMMAND_RELEASE_STREAM && r.CommandName != AMF0_COMMAND_FC_PUBLISH && r.CommandName != AMF0_COMMAND_UNPUBLISH {
		names := []string{AMF0_COMMAND_RELEASE_STREAM, AMF0_COMMAND_FC_PUBLISH, AMF0_COMMAND_UNPUBLISH}
		return Error{code: ERROR_RTMP_AMF0_DECODE, desc: fmt.Sprintf("amf0 decode name failed. expect=(%v), actual=%v", strings.Join(names, ","), r.CommandName)}
	}

	if r.TransactionId, err = codec.ReadNumber(); err != nil {
		return
	}
	if err = r.CommandObject.Read(codec); err != nil {
		return
	}
	if r.StreamName, err = codec.ReadString(); err != nil {
		return
	}

	return
}

/**
* response for SrsFMLEStartPacket.
 */
// @see: SrsFMLEStartResPacket
type FMLEStartResPacket struct {
	CommandName   string
	TransactionId float64
	CommandObject *Amf0Any // Null
	Args          *Amf0Any // Undefined

}

func NewFMLEStartResPacket(transaction_id float64) *FMLEStartResPacket {
	r := &FMLEStartResPacket{}
	r.CommandName = AMF0_COMMAND_RESULT
	r.TransactionId = transaction_id
	r.CommandObject = NewAmf0Null()
	r.Args = NewAmf0Undefined()
	return r
}

// Encoder
func (r *FMLEStartResPacket) GetPerferCid() (v int) {
	return RTMP_CID_OverConnection
}
func (r *FMLEStartResPacket) GetMessageType() (v byte) {
	return RTMP_MSG_AMF0CommandMessage
}
func (r *FMLEStartResPacket) GetSize() (v int) {
	return Amf0SizeString(r.CommandName) + Amf0SizeNumber() + r.CommandObject.Size() + r.Args.Size()
}
func (r *FMLEStartResPacket) Encode(s *Buffer) (err error) {
	codec := NewAmf0Codec(s)

	if err = codec.WriteString(r.CommandName); err != nil {
		return
	}
	if err = codec.WriteNumber(r.TransactionId); err != nil {
		return
	}
	if err = r.CommandObject.Write(codec); err != nil {
		return
	}
	if err = r.Args.Write(codec); err != nil {
		return
	}
	return
}

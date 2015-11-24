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

/****************************************************************************
*****************************************************************************
****************************************************************************/
/**
5. Protocol Control Messages
RTMP reserves message type IDs 1-7 for protocol control messages.
These messages contain information needed by the RTM Chunk Stream
protocol or RTMP itself. Protocol messages with IDs 1 & 2 are
reserved for usage with RTM Chunk Stream protocol. Protocol messages
with IDs 3-6 are reserved for usage of RTMP. Protocol message with ID
7 is used between edge server and origin server.
*/
const RTMP_MSG_SetChunkSize  = 0x01
const RTMP_MSG_AbortMessage  = 0x02
const RTMP_MSG_Acknowledgement  = 0x03
const RTMP_MSG_UserControlMessage  = 0x04
const RTMP_MSG_WindowAcknowledgementSize  = 0x05
const RTMP_MSG_SetPeerBandwidth  = 0x06
const RTMP_MSG_EdgeAndOriginServerCommand  = 0x07
/**
3. Types of messages
The server and the client send messages over the network to
communicate with each other. The messages can be of any type which
includes audio messages, video messages, command messages, shared
object messages, data messages, and user control messages.
3.1. Command message
Command messages carry the AMF-encoded commands between the client
and the server. These messages have been assigned message type value
of 20 for AMF0 encoding and message type value of 17 for AMF3
encoding. These messages are sent to perform some operations like
connect, createStream, publish, play, pause on the peer. Command
messages like onstatus, result etc. are used to inform the sender
about the status of the requested commands. A command message
consists of command name, transaction ID, and command object that
contains related parameters. A client or a server can request Remote
Procedure Calls (RPC) over streams that are communicated using the
command messages to the peer.
*/
const RTMP_MSG_AMF3CommandMessage = 17 // = 0x11
const RTMP_MSG_AMF0CommandMessage = 20 // = 0x14
/**
3.2. Data message
The client or the server sends this message to send Metadata or any
user data to the peer. Metadata includes details about the
data(audio, video etc.) like creation time, duration, theme and so
on. These messages have been assigned message type value of 18 for
AMF0 and message type value of 15 for AMF3.
*/
const RTMP_MSG_AMF0DataMessage = 18 // = 0x12
const RTMP_MSG_AMF3DataMessage = 15 // = 0x0F
/**
3.3. Shared object message
A shared object is a Flash object (a collection of name value pairs)
that are in synchronization across multiple clients, instances, and
so on. The message types kMsgContainer=19 for AMF0 and
kMsgContainerEx=16 for AMF3 are reserved for shared object events.
Each message can contain multiple events.
*/
const RTMP_MSG_AMF3SharedObject = 16 // = 0x10
const RTMP_MSG_AMF0SharedObject = 19 // = 0x13
/**
3.4. Audio message
The client or the server sends this message to send audio data to the
peer. The message type value of 8 is reserved for audio messages.
*/
const RTMP_MSG_AudioMessage = 8 // = 0x08
/* *
3.5. Video message
The client or the server sends this message to send video data to the
peer. The message type value of 9 is reserved for video messages.
These messages are large and can delay the sending of other type of
messages. To avoid such a situation, the video message is assigned
the lowest priority.
*/
const RTMP_MSG_VideoMessage = 9 // = 0x09
/**
3.6. Aggregate message
An aggregate message is a single message that contains a list of submessages.
The message type value of 22 is reserved for aggregate
messages.
*/
const RTMP_MSG_AggregateMessage = 22 // = 0x16
/****************************************************************************
*****************************************************************************
****************************************************************************/
/**
* 6.1.2. Chunk Message Header
* There are four different formats for the chunk message header,
* selected by the "fmt" field in the chunk basic header.
*/
// 6.1.2.1. Type 0
// Chunks of Type 0 are 11 bytes long. This type MUST be used at the
// start of a chunk stream, and whenever the stream timestamp goes
// backward (e.g., because of a backward seek).
const RTMP_FMT_TYPE0 = 0
// 6.1.2.2. Type 1
// Chunks of Type 1 are 7 bytes long. The message stream ID is not
// included; this chunk takes the same stream ID as the preceding chunk.
// Streams with variable-sized messages (for example, many video
// formats) SHOULD use this format for the first chunk of each new
// message after the first.
const RTMP_FMT_TYPE1 =  1
// 6.1.2.3. Type 2
// Chunks of Type 2 are 3 bytes long. Neither the stream ID nor the
// message length is included; this chunk has the same stream ID and
// message length as the preceding chunk. Streams with constant-sized
// messages (for example, some audio and data formats) SHOULD use this
// format for the first chunk of each message after the first.
const RTMP_FMT_TYPE2 = 2
// 6.1.2.4. Type 3
// Chunks of Type 3 have no header. Stream ID, message length and
// timestamp delta are not present; chunks of this type take values from
// the preceding chunk. When a single message is split into chunks, all
// chunks of a message except the first one, SHOULD use this type. Refer
// to example 2 in section 6.2.2. Stream consisting of messages of
// exactly the same size, stream ID and spacing in time SHOULD use this
// type for all chunks after chunk of Type 2. Refer to example 1 in
// section 6.2.1. If the delta between the first message and the second
// message is same as the time stamp of first message, then chunk of
// type 3 would immediately follow the chunk of type 0 as there is no
// need for a chunk of type 2 to register the delta. If Type 3 chunk
// follows a Type 0 chunk, then timestamp delta for this Type 3 chunk is
// the same as the timestamp of Type 0 chunk.
const RTMP_FMT_TYPE3 = 3

/****************************************************************************
*****************************************************************************
****************************************************************************/
/**
* 6. Chunking
* The chunk size is configurable. It can be set using a control
* message(Set Chunk Size) as described in section 7.1. The maximum
* chunk size can be 65536 bytes and minimum 128 bytes. Larger values
* reduce CPU usage, but also commit to larger writes that can delay
* other content on lower bandwidth connections. Smaller chunks are not
* good for high-bit rate streaming. Chunk size is maintained
* independently for each direction.
*/
const RTMP_DEFAULT_CHUNK_SIZE = 128
const RTMP_MIN_CHUNK_SIZE = 128
const RTMP_MAX_CHUNK_SIZE = 65536

/**
* 6.1. Chunk Format
* Extended timestamp: 0 or 4 bytes
* This field MUST be sent when the normal timsestamp is set to
* = 0xffffff, it MUST NOT be sent if the normal timestamp is set to
* anything else. So for values less than = 0xffffff the normal
* timestamp field SHOULD be used in which case the extended timestamp
* MUST NOT be present. For values greater than or equal to = 0xffffff
* the normal timestamp field MUST NOT be used and MUST be set to
* = 0xffffff and the extended timestamp MUST be sent.
*/
const RTMP_EXTENDED_TIMESTAMP  = 0xFFFFFF

/****************************************************************************
*****************************************************************************
****************************************************************************/
/**
* amf0 command message, command name macros
*/
const AMF0_COMMAND_CONNECT = "connect"
const AMF0_COMMAND_CREATE_STREAM = "createStream"
const AMF0_COMMAND_CLOSE_STREAM = "closeStream"
const AMF0_COMMAND_PLAY = "play"
const AMF0_COMMAND_PAUSE = "pause"
const AMF0_COMMAND_ON_BW_DONE = "onBWDone"
const AMF0_COMMAND_ON_STATUS = "onStatus"
const AMF0_COMMAND_RESULT = "_result"
const AMF0_COMMAND_ERROR = "_error"
const AMF0_COMMAND_RELEASE_STREAM = "releaseStream"
const AMF0_COMMAND_FC_PUBLISH = "FCPublish"
const AMF0_COMMAND_UNPUBLISH = "FCUnpublish"
const AMF0_COMMAND_PUBLISH = "publish"
const AMF0_DATA_SAMPLE_ACCESS = "|RtmpSampleAccess"
const AMF0_DATA_SET_DATAFRAME = "@setDataFrame"
const AMF0_DATA_ON_METADATA = "onMetaData"

/**
* band width check method name, which will be invoked by client.
* band width check mothods use SrsBandwidthPacket as its internal packet type,
* so ensure you set command name when you use it.
*/
// server play control
const SRS_BW_CHECK_START_PLAY = "onSrsBandCheckStartPlayBytes"
const SRS_BW_CHECK_STARTING_PLAY = "onSrsBandCheckStartingPlayBytes"
const SRS_BW_CHECK_STOP_PLAY = "onSrsBandCheckStopPlayBytes"
const SRS_BW_CHECK_STOPPED_PLAY = "onSrsBandCheckStoppedPlayBytes"

// server publish control
const SRS_BW_CHECK_START_PUBLISH  = "onSrsBandCheckStartPublishBytes"
const SRS_BW_CHECK_STARTING_PUBLISH = "onSrsBandCheckStartingPublishBytes"
const SRS_BW_CHECK_STOP_PUBLISH = "onSrsBandCheckStopPublishBytes"
const SRS_BW_CHECK_STOPPED_PUBLISH = "onSrsBandCheckStoppedPublishBytes"

// EOF control.
const SRS_BW_CHECK_FINISHED = "onSrsBandCheckFinished"
// for flash, it will sendout a final call,
// used to confirm got the report.
// actually, client send out this packet and close the connection,
// so server may cannot got this packet, ignore is ok.
const SRS_BW_CHECK_FLASH_FINAL = "finalClientPacket"

// client only
const SRS_BW_CHECK_PLAYING = "onSrsBandCheckPlaying"
const SRS_BW_CHECK_PUBLISHING = "onSrsBandCheckPublishing"

/****************************************************************************
*****************************************************************************
****************************************************************************/
/**
* the chunk stream id used for some under-layer message,
* for example, the PC(protocol control) message.
*/
const RTMP_CID_ProtocolControl = 0x02
/**
* the AMF0/AMF3 command message, invoke method and return the result, over NetConnection.
* generally use = 0x03.
*/
const RTMP_CID_OverConnection = 0x03
/**
* the AMF0/AMF3 command message, invoke method and return the result, over NetConnection,
* the midst state(we guess).
* rarely used, e.g. onStatus(NetStream.Play.Reset).
*/
const RTMP_CID_OverConnection2 = 0x04
/**
* the stream message(amf0/amf3), over NetStream.
* generally use = 0x05.
*/
const RTMP_CID_OverStream = 0x05
/**
* the stream message(amf0/amf3), over NetStream, the midst state(we guess).
* rarely used, e.g. play("mp4:mystram.f4v")
*/
const RTMP_CID_OverStream2 = 0x08
/**
* the stream message(video), over NetStream
* generally use = 0x06.
*/
const RTMP_CID_Video = 0x06
/**
* the stream message(audio), over NetStream.
* generally use = 0x07.
*/
const RTMP_CID_Audio = 0x07

/****************************************************************************
*****************************************************************************
****************************************************************************/

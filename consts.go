package main

// the following is the timeout for rtmp protocol,
// to avoid death connection.

// when got a messae header, there must be some data,
// increase recv timeout to got an entire message.
const SRS_MIN_RECV_TIMEOUT_MS = 60 * 1000

// the timeout to wait for client control message,
// if timeout, we generally ignore and send the data to client,
// generally, it's the pulse time for data seding.
const SRS_PULSE_TIMEOUT_MS = 200

// the pprof timeout, the ping message recv/send timeout
const SRS_PPROF_TIMEOUT_MS = 10 * 60 * 1000
const SRS_PPROF_PULSE_MS = 800
const SRS_PPROF_VHOST = "pprof"

// the timeout to wait client data,
// if timeout, close the connection.
const SRS_SEND_TIMEOUT_MS = 30 * 1000

// the timeout to send data to client,
// if timeout, close the connection.
const SRS_RECV_TIMEOUT_MS = 30 * 1000

// the timeout to wait client data, when client paused
// if timeout, close the connection.
const SRS_PAUSED_SEND_TIMEOUT_MS = 30 * 60 * 1000

// the timeout to send data to client, when client paused
// if timeout, close the connection.
const SRS_PAUSED_RECV_TIMEOUT_MS = 30 * 60 * 1000

// when stream is busy, for example, streaming is already
// publishing, when a new client to request to publish,
// sleep a while and close the connection.
const SRS_STREAM_BUSY_SLEEP_MS = 3 * 1000

// when error, forwarder sleep for a while and retry.
const SRS_FORWARDER_SLEEP_MS = 3 * 1000

// when error, encoder sleep for a while and retry.
const SRS_ENCODER_SLEEP_MS = 3 * 1000

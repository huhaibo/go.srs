package main

import "fmt"

// system control message,
// not an error, but special control logic.
// sys ctl: rtmp close stream, support replay.
const ERROR_CONTROL_RTMP_CLOSE = 100

/**
* whether the error code is an system control error.
 */
// @see: srs_is_system_control_error
func IsSystemControlError(err error) bool {
	if re, ok := err.(SrsError); ok {
		switch re.code {
		case ERROR_CONTROL_RTMP_CLOSE:
			return true
		}
	}
	return false
}

func IsSystemControlRtmpClose(err error) bool {
	if re, ok := err.(SrsError); ok {
		return re.code == ERROR_CONTROL_RTMP_CLOSE
	}
	return false
}

type SrsError struct {
	code int
	desc string
}

func (err SrsError) Error() string {
	return fmt.Sprintf("srs error code=%v: %s", err.code, err.desc)
}

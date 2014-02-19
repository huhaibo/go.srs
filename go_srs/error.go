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

import "fmt"

// system control message,
// not an error, but special control logic.
// sys ctl: rtmp close stream, support replay.
const ERROR_CONTROL_RTMP_CLOSE = 100

/**
* whether the error code is an system control error.
*/
// @see: srs_is_system_control_error
func IsSystemControlError(err error) (bool) {
	if re, ok := err.(SrsError); ok {
		switch re.code {
		case ERROR_CONTROL_RTMP_CLOSE:
			return true
		}
	}
	return false
}

func IsSystemControlRtmpClose(err error) (bool) {
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

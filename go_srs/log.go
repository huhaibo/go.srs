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

import (
	"fmt"
	"time"
)

/**
* id for log to identify the current client.
 */
type SrsLogId uint64
type SrsLogIdGetter interface {
	GetId() (SrsLogId)
}

/**
* tag for log to identigy the log tag
 */
type SrsLogTag string
type SrsLogTagGetter interface {
	GetTag() (SrsLogTag)
}

func SrsFatal(id SrsLogIdGetter, tag SrsLogTagGetter, format string, a ...interface{}) {
	fmt.Printf(fmt.Sprintf("[Fatal][%v][%v][%v]%v\n", time.Now().Format("2006-01-02 15:04:05"), id.GetId(), tag.GetTag(), format), a...)
}
func SrsWarn(id SrsLogIdGetter, tag SrsLogTagGetter, format string, a ...interface{}) {
	fmt.Printf(fmt.Sprintf("[Warn0][%v][%v][%v]%v\n", time.Now().Format("2006-01-02 15:04:05"), id.GetId(), tag.GetTag(), format), a...)
}
func SrsTrace(id SrsLogIdGetter, tag SrsLogTagGetter, format string, a ...interface{}) {
	fmt.Printf(fmt.Sprintf("[Trace][%v][%v][%v]%v\n", time.Now().Format("2006-01-02 15:04:05"), id.GetId(), tag.GetTag(), format), a...)
}
func SrsInfo(id SrsLogIdGetter, tag SrsLogTagGetter, format string, a ...interface{}) {
	return
	fmt.Printf(fmt.Sprintf("[Info0][%v][%v][%v]%v\n", time.Now().Format("2006-01-02 15:04:05"), id.GetId(), tag.GetTag(), format), a...)
}
func SrsVerbose(id SrsLogIdGetter, tag SrsLogTagGetter, format string, a ...interface{}) {
	return
	fmt.Printf(fmt.Sprintf("[Verbs][%v][%v][%v]%v\n", time.Now().Format("2006-01-02 15:04:05"), id.GetId(), tag.GetTag(), format), a...)
}

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
)

// AMF0 marker
const AMF0_Number = 0x00
const AMF0_Boolean = 0x01
const AMF0_String = 0x02
const AMF0_Object = 0x03
const AMF0_MovieClip = 0x04 // reserved, not supported
const AMF0_Null = 0x05
const AMF0_Undefined = 0x06
const AMF0_Reference = 0x07
const AMF0_EcmaArray = 0x08
const AMF0_ObjectEnd = 0x09
const AMF0_StrictArray = 0x0A
const AMF0_Date = 0x0B
const AMF0_LongString = 0x0C
const AMF0_UnSupported = 0x0D
const AMF0_RecordSet = 0x0E // reserved, not supported
const AMF0_XmlDocument = 0x0F
const AMF0_TypedObject = 0x10
// AVM+ object is the AMF3 object.
const AMF0_AVMplusObject = 0x11
// origin array whos data takes the same form as LengthValueBytes
const AMF0_OriginStrictArray = 0x20

// User defined
const AMF0_Invalid = 0x3F

/**
* to ensure in inserted order.
* for the FMLE will crash when AMF0Object is not ordered by inserted,
* if ordered in map, the string compare order, the FMLE will creash when
* get the response of connect app.
*/
// @see: SrsUnSortedHashtable
type Amf0UnSortedHashtable struct {
	property_index []string
	properties map[string]*Amf0Any
}
func NewAmf0UnSortedHashtable() (*Amf0UnSortedHashtable) {
	r := &Amf0UnSortedHashtable{}
	r.properties = make(map[string]*Amf0Any)
	return r
}
func (r *Amf0UnSortedHashtable) Count() (n int) {
	return len(r.properties)
}
func (r *Amf0UnSortedHashtable) Size() (n int) {
	if r.Count() <= 0 {
		return 0
	}
	for k, v := range r.properties {
		n += Amf0SizeUtf8(k)
		n += v.Size()
	}
	return
}
func (r *Amf0UnSortedHashtable) Write(codec *Amf0Codec) (err error) {
	// properties
	for _, k := range r.property_index {
		v := r.properties[k]

		if err = codec.WriteUtf8(k); err != nil {
			return
		}
		if err = v.Write(codec); err != nil {
			return
		}
	}
	return
}
func (r *Amf0UnSortedHashtable) Set(k string, v *Amf0Any) (err error) {
	if v == nil {
		err = Error{code:ERROR_GO_AMF0_NIL_PROPERTY, desc:"AMF0 object property value should never be nil"}
		return
	}

	if _, ok := r.properties[k]; !ok {
		r.property_index = append(r.property_index, k)
	}
	r.properties[k] = v
	return
}
func (r *Amf0UnSortedHashtable) GetPropertyString(k string) (v string, ok bool) {
	var prop *Amf0Any
	if prop, ok = r.properties[k]; !ok {
		return
	}
	return prop.String()
}
func (r *Amf0UnSortedHashtable) GetPropertyNumber(k string) (v float64, ok bool) {
	var prop *Amf0Any
	if prop, ok = r.properties[k]; !ok {
		return
	}
	return prop.Number()
}

/**
* 2.5 Object Type
* anonymous-object-type = object-marker *(object-property)
* object-property = (UTF-8 value-type) | (UTF-8-empty object-end-marker)
*/
// @see: SrsAmf0Object
type Amf0Object struct {
	marker byte
	properties *Amf0UnSortedHashtable
}
func NewAmf0Object() (*Amf0Object) {
	r := &Amf0Object{}
	r.marker = AMF0_Object
	r.properties = NewAmf0UnSortedHashtable()
	return r
}

func (r *Amf0Object) Size() (n int) {
	if n = r.properties.Size(); n <= 0 {
		return 0
	}

	n += 1
	n += Amf0SizeObjectEOF()
	return
}
func (r *Amf0Object) Read(codec *Amf0Codec) (err error) {
	// marker
	if !codec.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 object requires 1bytes marker"}
		return
	}

	if r.marker = codec.stream.ReadByte(); r.marker != AMF0_Object{
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 object marker invalid"}
		return
	}

	for !codec.stream.Empty() {
		// property-name: utf8 string
		var property_name string
		if property_name, err = codec.ReadUtf8(); err != nil {
			return
		}

		// property-value: any
		var property_value Amf0Any
		if err = property_value.Read(codec); err != nil {
			return
		}

		// AMF0 Object EOF.
		if len(property_name) <= 0 || property_value.IsNil() || property_value.IsObjectEof() {
			break
		}

		// add property
		if err = r.Set(property_name, &property_value); err != nil {
			return
		}
	}
	return
}
func (r *Amf0Object) Write(codec *Amf0Codec) (err error) {
	// marker
	if !codec.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write object marker failed"}
		return
	}
	codec.stream.WriteByte(byte(AMF0_Object))

	// properties
	if err = r.properties.Write(codec); err != nil {
		return
	}

	// object EOF
	return codec.WriteObjectEOF()
}
func (r *Amf0Object) Set(k string, v *Amf0Any) (err error) {
	return r.properties.Set(k, v)
}
func (r *Amf0Object) GetPropertyString(k string) (v string, ok bool) {
	return r.properties.GetPropertyString(k)
}
func (r *Amf0Object) GetPropertyNumber(k string) (v float64, ok bool) {
	return r.properties.GetPropertyNumber(k)
}

/**
* 2.10 ECMA Array Type
* ecma-array-type = associative-count *(object-property)
* associative-count = U32
* object-property = (UTF-8 value-type) | (UTF-8-empty object-end-marker)
*/
// @see: SrsASrsAmf0EcmaArray
type Amf0EcmaArray struct {
	marker byte
	count uint32
	properties *Amf0UnSortedHashtable
}
func NewAmf0EcmaArray() (*Amf0EcmaArray) {
	r := &Amf0EcmaArray{}
	r.marker = AMF0_EcmaArray
	r.properties = NewAmf0UnSortedHashtable()
	return r
}

func (r *Amf0EcmaArray) Size() (n int) {
	if n = r.properties.Size(); n <= 0 {
		return 0
	}

	n += 1
	n += 4
	n += Amf0SizeObjectEOF()
	return
}
// srs_amf0_read_ecma_array
func (r *Amf0EcmaArray) Read(codec *Amf0Codec) (err error) {
	// marker
	if !codec.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 EcmaArray requires 1bytes marker"}
		return
	}

	if r.marker = codec.stream.ReadByte(); r.marker != AMF0_EcmaArray{
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 EcmaArray marker invalid"}
		return
	}

	// count
	if !codec.stream.Requires(4) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 read ecma_array count failed"}
		return
	}
	r.count = codec.stream.ReadUInt32()

	for !codec.stream.Empty() {
		// property-name: utf8 string
		var property_name string
		if property_name, err = codec.ReadUtf8(); err != nil {
			return
		}

		// property-value: any
		var property_value Amf0Any
		if err = property_value.Read(codec); err != nil {
			return
		}

		// AMF0 Object EOF.
		if len(property_name) <= 0 || property_value.IsNil() || property_value.IsObjectEof() {
			break
		}

		// add property
		if err = r.Set(property_name, &property_value); err != nil {
			return
		}
	}
	return
}
// srs_amf0_write_ecma_array
func (r *Amf0EcmaArray) Write(codec *Amf0Codec) (err error) {
	// marker
	if !codec.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write EcmaArray marker failed"}
		return
	}
	codec.stream.WriteByte(byte(AMF0_EcmaArray))

	// count
	if !codec.stream.Requires(4) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write ecma_array count failed"}
		return
	}
	codec.stream.WriteUInt32(r.count)

	// properties
	if err = r.properties.Write(codec); err != nil {
		return
	}

	// object EOF
	return codec.WriteObjectEOF()
}
func (r *Amf0EcmaArray) Set(k string, v *Amf0Any) (err error) {
	err = r.properties.Set(k, v)
	r.count = uint32(r.properties.Count())
	return
}
func (r *Amf0EcmaArray) GetPropertyString(k string) (v string, ok bool) {
	return r.properties.GetPropertyString(k)
}
func (r *Amf0EcmaArray) GetPropertyNumber(k string) (v float64, ok bool) {
	return r.properties.GetPropertyNumber(k)
}

/**
* any amf0 value.
* 2.1 Types Overview
* value-type = number-type | boolean-type | string-type | object-type
* 		| null-marker | undefined-marker | reference-type | ecma-array-type
* 		| strict-array-type | date-type | long-string-type | xml-document-type
* 		| typed-object-type
* create any with NewAmf0(), or create a default one and Read from stream.
*/
// @see: SrsAmf0Any
type Amf0Any struct {
	Marker byte
	Value interface {}
}
func NewAmf0(v interface {}) (*Amf0Any) {
	switch t := v.(type) {
	case bool:
		return &Amf0Any{ Marker:AMF0_Boolean, Value:t }
	case string:
		return &Amf0Any{ Marker:AMF0_String, Value:t }
	case int:
		return &Amf0Any{ Marker:AMF0_Number, Value:float64(t) }
	case float64:
		return &Amf0Any{ Marker:AMF0_Number, Value:t }
	case *Amf0Object:
		return &Amf0Any{ Marker:AMF0_Object, Value:t }
	case *Amf0EcmaArray:
		return &Amf0Any{ Marker:AMF0_EcmaArray, Value:t }
	}
	return nil
}
func NewAmf0Null() (*Amf0Any) {
	return &Amf0Any{ Marker:AMF0_Null }
}
func NewAmf0Undefined() (*Amf0Any) {
	return &Amf0Any{ Marker:AMF0_Undefined }
}
func (r *Amf0Any) Size() (int) {
	switch {
	case r.Marker == AMF0_String:
		v, _ := r.String()
		return Amf0SizeString(v)
	case r.Marker == AMF0_Boolean:
		return Amf0SizeBoolean()
	case r.Marker == AMF0_Number:
		return Amf0SizeNumber()
	case r.Marker == AMF0_Null || r.Marker == AMF0_Undefined:
		return Amf0SizeNullOrUndefined()
	case r.Marker == AMF0_ObjectEnd:
		return Amf0SizeObjectEOF()
	case r.Marker == AMF0_Object:
		v, _ := r.Object()
		return v.Size()
	case r.Marker == AMF0_EcmaArray:
		v, _ := r.EcmaArray()
		return v.Size()
		// TODO: FIXME: implements it.
	}
	return 0
}
func (r *Amf0Any) Write(codec *Amf0Codec) (err error) {
	switch {
	case r.Marker == AMF0_String:
		v, _ := r.String()
		return codec.WriteString(v)
	case r.Marker == AMF0_Boolean:
		v, _ := r.Boolean()
		return codec.WriteBoolean(v)
	case r.Marker == AMF0_Number:
		v, _ := r.Number()
		return codec.WriteNumber(v)
	case r.Marker == AMF0_Null:
		return codec.WriteNull()
	case r.Marker == AMF0_Undefined:
		return codec.WriteUndefined()
	case r.Marker == AMF0_ObjectEnd:
		return codec.WriteObjectEOF()
	case r.Marker == AMF0_Object:
		v, _ := r.Object()
		return v.Write(codec)
	case r.Marker == AMF0_EcmaArray:
		v, _ := r.EcmaArray()
		return v.Write(codec)
		// TODO: FIXME: implements it.
	}
	return
}
func (r *Amf0Any) Read(codec *Amf0Codec) (err error) {
	// marker
	if !codec.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 any requires 1bytes marker"}
		return
	}
	r.Marker = codec.stream.ReadByte()
	codec.stream.Skip(-1)

	switch {
	case r.Marker == AMF0_String:
		r.Value, err = codec.ReadString()
	case r.Marker == AMF0_Boolean:
		r.Value, err = codec.ReadBoolean()
	case r.Marker == AMF0_Number:
		r.Value, err = codec.ReadNumber()
	case r.Marker == AMF0_Null || r.Marker == AMF0_Undefined || r.Marker == AMF0_ObjectEnd:
		codec.stream.ReadByte()
	case r.Marker == AMF0_Object:
		r.Value, err = codec.ReadObject()
	case r.Marker == AMF0_EcmaArray:
		r.Value, err = codec.ReadEcmaArray()
	// TODO: FIXME: implements it.
	default:
		err = Error{code:ERROR_RTMP_AMF0_INVALID, desc:fmt.Sprintf("invalid amf0 message type. marker=%#x", r.Marker)}
	}

	return
}
func (r *Amf0Any) IsNil() (v bool) {
	return r.Value == nil
}
func (r *Amf0Any) IsObjectEof() (v bool) {
	return r.Marker == AMF0_ObjectEnd
}
func (r *Amf0Any) Object() (v *Amf0Object, ok bool) {
	if r.Marker == AMF0_Object {
		v, ok = r.Value.(*Amf0Object), true
	}
	return
}
func (r *Amf0Any) EcmaArray() (v *Amf0EcmaArray, ok bool) {
	if r.Marker == AMF0_EcmaArray {
		v, ok = r.Value.(*Amf0EcmaArray), true
	}
	return
}
func (r *Amf0Any) String() (v string, ok bool) {
	if r.Marker == AMF0_String {
		v, ok = r.Value.(string), true
	}
	return
}
func (r *Amf0Any) Number() (v float64, ok bool) {
	if r.Marker == AMF0_Number {
		v, ok = r.Value.(float64), true
	}
	return
}
func (r *Amf0Any) Boolean() (v bool, ok bool) {
	if r.Marker == AMF0_Boolean {
		v, ok = r.Value.(bool), true
	}
	return
}

type Amf0Codec struct {
	stream *Buffer
}
func NewAmf0Codec(stream *Buffer) (*Amf0Codec) {
	r := Amf0Codec{}
	r.stream = stream
	return &r
}

// Size
func Amf0SizeString(v string) (int) {
	return 1 + Amf0SizeUtf8(v)
}
func Amf0SizeUtf8(v string) (int) {
	return 2 + len(v)
}
func Amf0SizeNumber() (int) {
	return 1 + 8
}
func Amf0SizeNullOrUndefined() (int) {
	return 1
}
func Amf0SizeBoolean() (int) {
	return 1 + 1
}
func Amf0SizeObjectEOF() (int) {
	return 2 + 1
}

// srs_amf0_read_string
func (r *Amf0Codec) ReadString() (v string, err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 string requires 1bytes marker"}
		return
	}

	if marker := r.stream.ReadByte(); marker != AMF0_String {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 string marker invalid"}
		return
	}

	v, err = r.ReadUtf8()
	return
}
// srs_amf0_write_string
func (r *Amf0Codec) WriteString(v string) (err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write string marker failed"}
		return
	}
	r.stream.WriteByte(byte(AMF0_String))
	return r.WriteUtf8(v)
}
// srs_amf0_write_boolean
func (r *Amf0Codec) WriteBoolean(v bool) (err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write bool marker failed"}
		return
	}
	r.stream.WriteByte(byte(AMF0_Boolean))

	// value
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write bool value failed"}
		return
	}
	if v {
		r.stream.WriteByte(byte(0x01))
	} else {
		r.stream.WriteByte(byte(0x00))
	}
	return
}
// srs_amf0_read_utf8
func (r *Amf0Codec) ReadUtf8() (v string, err error) {
	// len
	if !r.stream.Requires(2) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 utf8 len requires 2bytes"}
		return
	}
	len := r.stream.ReadUInt16()

	// empty string
	if len <= 0 {
		return
	}

	// data
	if !r.stream.Requires(int(len)) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 utf8 data requires more bytes"}
		return
	}
	v = string(r.stream.Read(int(len)))

	// support utf8-1 only
	// 1.3.1 Strings and UTF-8
	// UTF8-1 = %x00-7F
	for _, ch := range v {
		if (ch & 0x80) != 0 {
			// ignored. only support utf8-1, 0x00-0x7F
			//err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"only support utf8-1, 0x00-0x7F"}
			//return
		}
	}

	return
}
// srs_amf0_write_utf8
func (r *Amf0Codec) WriteUtf8(v string) (err error) {
	// len
	if !r.stream.Requires(2) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write string length failed"}
		return
	}
	r.stream.WriteUInt16(uint16(len(v)))

	// empty string
	if len(v) <= 0 {
		return
	}

	// data
	if !r.stream.Requires(len(v)) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write string data failed"}
		return
	}
	r.stream.Write([]byte(v))
	return
}
// srs_amf0_read_number
func (r *Amf0Codec) ReadNumber() (v float64, err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 number requires 1bytes marker"}
		return
	}

	if marker := r.stream.ReadByte(); marker != AMF0_Number{
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 number marker invalid"}
		return
	}

	// value
	if !r.stream.Requires(8) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 number requires 8bytes value"}
		return
	}
	v = r.stream.ReadFloat64()

	return
}
// srs_amf0_write_number
func (r *Amf0Codec) WriteNumber(v float64) (err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write number marker failed"}
		return
	}
	r.stream.WriteByte(byte(AMF0_Number))

	// value
	if !r.stream.Requires(8) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write number value failed"}
		return
	}
	r.stream.WriteFloat64(v)

	return
}
// srs_amf0_write_null
func (r *Amf0Codec) WriteNull() (err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write null marker failed"}
		return
	}
	r.stream.WriteByte(byte(AMF0_Null))

	return
}
// srs_amf0_read_null
func (r *Amf0Codec) ReadNull() (err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 read null marker failed"}
		return
	}
	r.stream.ReadByte()

	return

}
// srs_amf0_read_undefined
func (r *Amf0Codec) WriteUndefined() (err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write undefined marker failed"}
		return
	}
	r.stream.WriteByte(byte(AMF0_Undefined))

	return
}
// srs_amf0_read_boolean
func (r *Amf0Codec) ReadBoolean() (v bool, err error) {
	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 bool requires 1bytes marker"}
		return
	}

	if marker := r.stream.ReadByte(); marker != AMF0_Boolean{
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 bool marker invalid"}
		return
	}

	// value
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_DECODE, desc:"amf0 bool requires 8bytes value"}
		return
	}

	if r.stream.ReadByte() == 0 {
		v = false
	} else {
		v = true
	}

	return
}
// srs_amf0_read_object
func (r *Amf0Codec) ReadObject() (v *Amf0Object, err error) {
	// value
	v = NewAmf0Object()
	return v, v.Read(r)
}
// srs_amf0_read_ecma_array
func (r *Amf0Codec) ReadEcmaArray() (v *Amf0EcmaArray, err error) {
	// value
	v = NewAmf0EcmaArray()
	return v, v.Read(r)
}
// srs_amf0_write_object
func (r *Amf0Codec) WriteObject(v *Amf0Object) (err error) {
	return v.Write(r)
}
// srs_amf0_read_ecma_array
func (r *Amf0Codec) WriteEcmaArray(v *Amf0EcmaArray) (err error) {
	return v.Write(r)
}
// srs_amf0_write_object_eof
func (r *Amf0Codec) WriteObjectEOF() (err error) {
	// value
	if !r.stream.Requires(2) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write object eof value failed"}
		return
	}
	r.stream.WriteUInt16(uint16(0))

	// marker
	if !r.stream.Requires(1) {
		err = Error{code:ERROR_RTMP_AMF0_ENCODE, desc:"amf0 write object eof marker failed"}
		return
	}
	r.stream.WriteByte(byte(AMF0_ObjectEnd))
	return
}

package main

import (
	"fmt"
	"strconv"
	"math"
)

func main() {
	str := "abcdefg"
	bs := []byte(str)
	fmt.Println("string=>[]byte is", bs)

	bs = []byte{110, 111, 112, 113, 114, 115, 116, 117}
	str = string(bs)
	fmt.Println("[]byte=>string is", str)

	vi := int(100)
	str = strconv.Itoa(vi)
	fmt.Println("int=>str is", str)
	fmt.Println("int=>str is", fmt.Sprintf("%v", vi))

	vfloat64 := float64(3.0)
	vuint64 := math.Float64bits(vfloat64)
	fmt.Printf("float64=>uint64 is %#x\n", vuint64)

	vfloat64 = math.Float64frombits(vuint64)
	fmt.Println("uint64=>float64 is", vfloat64)
}

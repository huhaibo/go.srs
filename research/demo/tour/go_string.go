package main

import (
	"fmt"
	"reflect"
)

func main() {
	str := "中文\x80混合英文Hello世界"
	fmt.Println(str)

	for pos, ch := range str {
		fmt.Printf("byte pos=%v, char=%#U, type=%v\n", pos, ch, reflect.TypeOf(ch))
	}

	for i := 0; i < len(str); i++ {
		fmt.Printf("byte pos=%v, char=%#U\n", i, str[i])
	}
}

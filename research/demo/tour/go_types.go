package main

import (
	"fmt"
	"reflect"
)

func main() {
	var a interface {} = 10
	fmt.Println(a)
	fmt.Println(reflect.TypeOf(a))
	fmt.Println(a.(int) + 10)
	fmt.Println(reflect.ValueOf(a).Int() + 10)
}

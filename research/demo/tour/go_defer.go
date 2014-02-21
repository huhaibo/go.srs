package main

import (
	"fmt"
	"runtime"
)

func panic_f() {
	defer fmt.Println("panic occur, defered")
	panic(fmt.Sprintf("panic occur"))
}
func my_demo(count *int) (code int) {
	defer func(){
		code++
		if r := recover(); r != nil {
			fmt.Println("recover success in parent function")
			return
		}
		fmt.Println("recover failed in parent function")
		fmt.Print(runtime.Stack())
	}()

	code = 0
	*count++

	panic_f()

	return code
}

func main() {
	count := 1
	code := my_demo(&count)
	fmt.Printf("count=%v, code=%v\n", count, code)
}

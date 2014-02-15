package main

import (
	"fmt"
)

type MyError struct {
	code int
	desc string
}
func (e MyError) Error() (string) {
	return fmt.Sprintf("code=%v, desc: %v", e.code, e.desc)
}

func panic_error() {
	panic(MyError{code:100, desc:"system unknown error"})
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("detect from error", r)
			panic(r)
		}
	}()
	panic_error()
}

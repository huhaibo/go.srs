package main

import "fmt"

func ret_local_var() (int, int, bool, bool, bool) {
	var a, b int
	var c, d, e bool
	return a, b, c, d, e
}

func init_local_var() (int, bool) {
	var a = 100
	var b = true
	return a, b
}

func auto_local_var() (int, bool) {
	a, b := 110, true
	return a, b
}

func main() {
	a, b, c, d, e := ret_local_var()
	fmt.Printf("ret_local_var() (int, int, bool, bool, bool)\n\tret=%v,%v,%v,%v,%v\n", a, b, c, d, e)

	a, c = init_local_var()
	fmt.Printf("init_local_var()(int, bool)\n\tret=%v,%v\n", a, c)

	a, c = auto_local_var()
	fmt.Printf("auto_local_var()(int, bool)\n\tret=%v,%v\n", a, c)
}

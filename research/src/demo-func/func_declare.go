package main

import (
	"fmt"
)

func my_func(pa int) int {
	return pa * 10
}

func my_func_add(a int, b int) int {
	return a + b
}

func my_func_tree_add(a, b, c int) int {
	return a + b + c
}

func main(){
	pa := 10
	fmt.Printf("my_func(pa int) int\n\tpa=%d, ret=%d\n", pa, my_func(pa))

	a := 3
	b := 5
	fmt.Printf("my_func(a int, b int) int\n\ta=%d, b=%d, ret=%d\n", a, b, my_func_add(a, b))

	c := 9
	fmt.Printf("my_fnc(a, b, c int) int\n\ta=%d, b=%d, c=%d, ret=%d\n", a, b, c, my_func_tree_add(a, b, c))
}

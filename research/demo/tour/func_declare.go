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

func my_swap(a, b int) (int, int) {
	return b, a
}

func my_swap2(a, b int) (x, y int) {
	x = b;
	y = a;
	return;
}

func my_print(user int, params ...string) {
	fmt.Println("user:", user)
	fmt.Println("params:", params)
}

func main(){
	pa := 10
	fmt.Printf("my_func(pa int) int\n\tpa=%d, ret=%d\n", pa, my_func(pa))

	a := 3
	b := 5
	fmt.Printf("my_func(a int, b int) int\n\ta=%d, b=%d, ret=%d\n", a, b, my_func_add(a, b))

	c := 9
	fmt.Printf("my_fnc(a, b, c int) int\n\ta=%d, b=%d, c=%d, ret=%d\n", a, b, c, my_func_tree_add(a, b, c))

	pa, pb := my_swap(a, b)
	fmt.Printf("my_swap(a, b) (int, int)\n\ta=%d, b=%d, ret=%v,%v\n", a, b, pa, pb)

	pa, pb = my_swap2(a, b)
	fmt.Printf("my_swap2(a, b) (x, y, int)\n\ta=%v, b=%v, ret=%v,%v\n", a, b, pa, pb)

	fmt.Println("variant param keyword: ...")
	my_print(10, "a", "b", "c")
	my_print(11, []string{"a", "b", "c"}...)
	my_print(12, make([]string, 3)...)
}

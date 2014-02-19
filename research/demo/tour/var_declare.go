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

type Winlin struct {
	name string
	age int
}

func my_declare_ret() (n int) {
	n = 10
	if n, ok := 100, true; ok && n > 0 {
		return
	}
	return
}

func main() {
	// n is shadowed during return
	//fmt.Println(my_declare_ret())

	str := "abcdefg"
	fmt.Println([]byte(str))

	//no new variables on left side of :=
	//str := "new string"
	//fmt.Println(str)

	var winlin Winlin
	fmt.Println("var winlin is", winlin, ", ptr is", &winlin)

	a, b, c, d, e := ret_local_var()
	fmt.Printf("ret_local_var() (int, int, bool, bool, bool)\n\tret=%v,%v,%v,%v,%v\n", a, b, c, d, e)

	a, c = init_local_var()
	fmt.Printf("init_local_var()(int, bool)\n\tret=%v,%v\n", a, c)

	a, c = auto_local_var()
	fmt.Printf("auto_local_var()(int, bool)\n\tret=%v,%v\n", a, c)

	vi := 10
	fmt.Println("vi=", vi, ", init.")
	vi = 20
	fmt.Println("vi=", vi, ", finish.")

	wl := new(Winlin)
	wl.name = "winlin object"
	fmt.Println(wl)

	wl0 := &(*wl)
	wl0.name = "changed winlin object"
	fmt.Println(wl)
	fmt.Println(wl0)

	wl1 := *wl
	wl1.name = "changed by value"
	fmt.Println(wl)
	fmt.Println(wl1)
}

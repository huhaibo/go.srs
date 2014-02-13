package main

import (
	"fmt"
)

type Winlin struct {
	name string
	age int
	email string
}

func main() {
	if true {
		pi := new(int)
		fmt.Printf("new pi=%v, *pi=%v\n", pi, *pi)

		pw := new(Winlin)
		fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pw.name, pw.age, pw.email)

		pw = &Winlin{"winlin", 28, "winterserver@126.com"}
		fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pw.name, pw.age, pw.email)
		pin := *pw
		fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pin.name, pin.age, pin.email)

		pw = &Winlin{name:"winlin", age:29, email:"winterserver@126.com"}
		fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pw.name, pw.age, pw.email)

		const Enone,Eio = 1,0
		m := map[int]string {Enone: "no error", Eio: "Eio"}
		fmt.Println("m=", m)
		s := []string {Enone: "no error", Eio: "Eio"}
		fmt.Println("s=", s)

		var wlo Winlin
		fmt.Printf("winlin, name=%v, age=%v, email=%v\n", wlo.name, wlo.age, wlo.email)

		var slice_int []int
		for {
			if slice_int == nil {
				fmt.Println("slice int is nil")
			} else {
				fmt.Println("slice int is not nil")
				break
			}
			fmt.Println("make slice int")
			slice_int = make([]int, 10)
		}

		arr_int := [3]int{5, 8, 3}
		carr_int := arr_int // deep copy
		arr_int[2] = 1
		carr_int[2] = 2
		fmt.Println("array is deep copy")
		fmt.Println(arr_int)
		fmt.Println(carr_int)

		for i := 0; i < len(slice_int); i++ {
			slice_int[i] = 3 + i
		}
		fmt.Println("slice is analogous to pointer")
		cslice_int := slice_int
		slice_int[1] = 33
		cslice_int[1] = 34
		fmt.Println(slice_int)
		fmt.Println(cslice_int)

		fmt.Println("sub slice is analogous to pointer")
		cslice_int = slice_int[3:]
		cslice_int[1] = 45
		fmt.Println(slice_int[3:])
		fmt.Println(cslice_int)
	}

	if true {
		slice := make([]int, 10)
		fmt.Printf("slice len=%v, cap=%v\n", len(slice), cap(slice))
		slice = append(slice, 1)
		fmt.Printf("slice len=%v, cap=%v\n", len(slice), cap(slice))
		data := make([]int, 30)
		slice = append(slice, data...)
		fmt.Printf("slice len=%v, cap=%v\n", len(slice), cap(slice))
	}

	if true {
		fmt.Println("map key and value are any value")
		mp := map[interface{}] interface{} {
			10: "ok",
			"winlin": "hello",
		}
		fmt.Println(mp)
	}
	if true {
		users := map[string]string {
			"winlin": "beijing",
			"stone": "england",
		}
		user,ok := users["winlin"]
		fmt.Printf("user=%v,ok=%v\n", user, ok)
		var user1,ok1 = users["none"]
		fmt.Printf("user=%v,ok=%v\n", user1, ok1)
		user,ok = users["stone"]
		fmt.Printf("user=%v,ok=%v\n", user, ok)

		delete(users, "winlin")
		user,ok = users["winlin"]
		fmt.Printf("user=%v,ok=%v\n", user, ok)
	}
}


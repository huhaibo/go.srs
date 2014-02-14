package main

import (
	"fmt"
	"reflect"
)

type Winlin struct {
	name string
	age int
	email string
}
func (w Winlin) Create() (*Winlin) {
	p := new(Winlin)
	p.name = w.name
	p.age = w.age
	p.email = w.email
	return p
}
func (w *Winlin) String() (string) {
	return fmt.Sprintf("name:%v, age:%v, email:%v", w.name, w.age, w.email)
}

func main() {
	var a interface {} = 10
	fmt.Println(a)
	fmt.Println(reflect.TypeOf(a))
	fmt.Println(a.(int) + 10)
	fmt.Println(reflect.ValueOf(a).Int() + 10)

	pw := new(Winlin)
	fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pw.name, pw.age, pw.email)

	pw = &Winlin{"winlin", 28, "winterserver@126.com"}
	fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pw.name, pw.age, pw.email)
	pin := *pw
	fmt.Printf("winlin, name=%v, age=%v, email=%v\n", pin.name, pin.age, pin.email)

	fmt.Println(Winlin{}.Create())
	fmt.Println(pw.Create())
}

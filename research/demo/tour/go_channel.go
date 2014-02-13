package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("channel demo.")
	var my_go_routine = func(cmd chan int, data chan string){
		<- cmd
		fmt.Println("go routine run")
		time.Sleep(1 * time.Second)
		data <- "success"
		fmt.Println("go routine quit")
	}

	cmd := make(chan int)
	data := make(chan string)
	go my_go_routine(cmd, data)

	fmt.Println("run go routine in 3s")
	time.Sleep(3 * time.Second)

	cmd <- 0
	fmt.Println("cmd sent, start the go routine")

	v := <- data
	fmt.Println("exit, data is", v)
}

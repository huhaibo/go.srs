package main

import (
	"fmt"
	"time"
)

func control_channel() {
	fmt.Println("channel demo.")
	var my_go_routine = func(cmd chan int, data chan string){
		<- cmd
		fmt.Println("go routine run, quit in 3s")
		time.Sleep(3 * time.Second)
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

func close_channel() {
	ch := make(chan int)
	qc := make(chan int)

	go func() {
		for {
			v, ok := <- ch
			fmt.Println("get a value from channel:", v, ok)
			if !ok {
				time.Sleep(3 * time.Second)
			}
		}
		qc <- 1
	} ()

	time.Sleep(3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	fmt.Println("wait for goroutine to quit.")
	<- qc
}

func select_channel() {
	ch := make(chan int)
	qc := make(chan int)

	go func(){
		timeout := time.After(2 * time.Second)
		for {
			select{
			case v1,ok := <- ch:
				fmt.Println("got", v1, ok)
			case <- timeout:
				fmt.Println("tick")
			default:
				fmt.Printf("not ready\n")
				time.Sleep(3 * time.Second)
			}
		}
		qc <- 0
	}()

	time.Sleep(5 * time.Second)
	ch <- 1000

	<- qc
}

func main() {
	select_channel()
	return
	close_channel()
	control_channel()
}

package main

import (
	"fmt"
	"time"
	"runtime"
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

func close_write_channel() {
	ch := make(chan int, 3)

	pfun := func(id int){
		defer func(){
			if r := recover(); r != nil {
				if re, ok := r.(runtime.Error); ok {
					fmt.Println("runtime.Error:", re)
				}
				return
			}
		}()
		for {
			// if channel closed:
			// panic: runtime error: send on closed channel
			ch <- id
			time.Sleep(time.Duration(id) * time.Second)
		}
	}
	go pfun(1)
	go pfun(2)

	for i := 0; i < 3; i++ {
		v, ok := <-ch
		fmt.Println(v, ok)
	}
	close(ch)

	for {
		v, ok := <-ch
		fmt.Println(v, ok)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	close_write_channel()
	return
	select_channel()
	close_channel()
	control_channel()
}

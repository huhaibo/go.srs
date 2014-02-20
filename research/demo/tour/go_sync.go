package main

import (
	"fmt"
	"time"
	"sync"
	"runtime"
)

type Data struct {
	id int
	count int
}
func channel_share_fun(count int, ch chan int) {
	for i := 0; i < count; i ++ {
		if v, ok := <- ch; ok {
			ch <- v+1
		}
	}
}
func channel_share(nb_count int) {
	ch := make(chan int)
	c0, c1 := 10, nb_count
	for i := 0; i < c0; i++ {
		go channel_share_fun(c1, ch)
	}

	ch <- 0
	time.Sleep(5 * time.Second)
	if v,ok := <-ch; ok {
		fmt.Println("quit, v=", v, ",expect=", c0*c1, ", success=", v == c0*c1)
	}
}
func channel_share_2CPU(nb_count int) {
	runtime.GOMAXPROCS(2)
	channel_share(nb_count)
}

func defer_test_fun(id int, ch chan int) {
	defer func(){
		ch <- id
		fmt.Println("EOF")
	}()
}
func defer_test() {
	ch := make(chan int)

	for i := 0; i < 10; i++ {
		go defer_test_fun(i, ch)
	}
	for i := 0; i < 10; i++ {
		if v, ok := <- ch; ok {
			fmt.Println("got the exited chan:", v, ok)
		}
	}
	time.Sleep(3 * time.Second)
	fmt.Println("quit")
}

// it's ok, for the goroutine is single thread
/*
test the sync of go, to communicate by sharing memory
10 quit 87852670 , v= 175523153
10 quit 88378929 , v= 176231599
main vc= 176231599 ,vcp= 176231599
 */
func share_variable_fun(id int, qc chan int, data chan int, v *int) {
	vc := 0
	for {
		select {
		case <- qc:
			fmt.Println(id, "quit", vc, ", v=", *v)
			data <- vc
			return
		default:
			*v++
			vc++
		}
	}
}
func share_variable() {
	ch := make(chan int)
	data := make(chan int)
	vc := 0
	go share_variable_fun(10, ch, data, &vc)
	go share_variable_fun(10, ch, data, &vc)
	time.Sleep(5 * time.Second)
	ch <- 1
	ch <- 1
	time.Sleep(1 * time.Second)
	vcp0 := <- data
	vcp0 += <- data
	fmt.Println("main vc=", vc, ",vcp=", vcp0)
}

// it's ok, for the goroutine use mutex to sync
/**
test the sync of go, to communicate by sharing memory
10 use lock: quit 20000876 , v= 36555666
10 use lock: quit 16703988 , v= 36704864
use lock: main vc= 36704864 ,vcp= 36704864
 */
func share_variable_lock_fun_change(lock *sync.Mutex, vc *int, v *int) {
	lock.Lock()
	defer lock.Unlock()

	*v++
	*vc++
}
func share_variable_lock_fun(lock *sync.Mutex, id int, qc chan int, data chan int, v *int) {
	vc := 0
	for {
		select {
		case <- qc:
			fmt.Println(id, "use lock: quit", vc, ", v=", *v)
			data <- vc
			return
		default:
			share_variable_lock_fun_change(lock, &vc, v)
		}
	}
}
func share_variable_lock() {
	ch := make(chan int)
	data := make(chan int)
	lock := &sync.Mutex{}

	vc := 0
	go share_variable_lock_fun(lock, 10, ch, data, &vc)
	go share_variable_lock_fun(lock, 10, ch, data, &vc)
	time.Sleep(5 * time.Second)
	ch <- 1
	ch <- 1
	time.Sleep(1 * time.Second)
	vcp0 := <- data
	vcp0 += <- data
	fmt.Println("use lock: main vc=", vc, ",vcp=", vcp0)
}

/**
failed for 2thread need sync
test the sync of go, to communicate by sharing memory
10 quit 78346530 , v= 162976399
10 quit 84676800 , v= 163015173
main vc= 163015173 ,vcp= 163023330
 */
func share_variable_2CPU() {
	runtime.GOMAXPROCS(2)
	share_variable()
}

/**
ok for use lock.
test the sync of go, to communicate by sharing memory
10 use lock: quit 1490834 , v= 3120519
10 use lock: quit 1629685 , v= 3120519
use lock: main vc= 3120519 ,vcp= 3120519
 */
func share_variable_lock_2CPU() {
	runtime.GOMAXPROCS(2)
	share_variable_lock()
}

func main() {
	fmt.Println("test the sync of go, to communicate by sharing memory")
	//defer_test()
	//channel_share(2755245)
	//channel_share_2CPU(900)
	//share_variable()
	//share_variable_lock()
	//share_variable_2CPU()
	share_variable_lock_2CPU()
}

package main

import (
	"fmt"
	"time"
)

func main() {
	var fun = func (id int) {
		count := 0
		for {
			if (count % 1500000000) == 0 {
				fmt.Printf("[%v] id=%v, count=%v\n", time.Now().Format("2006-1-06 15:04:05"), id, count)
			}
			count++
		}
	}
	go fun(101)
	go fun(102)
	time.Sleep(300 * time.Second)
}

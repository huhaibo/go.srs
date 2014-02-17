package main

import "fmt"

const (
	_ = iota
	KB int = iota
	MB
	GB
)
func main() {
	fmt.Println("KB=", KB)
	fmt.Println("MB=", MB)
}

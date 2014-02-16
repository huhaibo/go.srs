package main

import "fmt"

func main() {
	var m = map[string]string {
		"vwx": "winlinx",
		"user1": "winlin",
		"user2": "hello",
	}
	fmt.Println(m)

	m = make(map[string]string)
	m["axxx"] = "axxxxxxxx"
	m["yzxx"] = "yzxxxxxx"
	m["mxxxxx"] = "mxxxxxx"
	fmt.Println(m)

	v, ok := m["axxx"]
	fmt.Println("v=", v, ",ok=", ok)
	v, ok = m["abc"]
	fmt.Println("v=", v, ",ok=", ok)

	delete(m, "axxx")
	v, ok = m["axxx"]
	fmt.Println("v=", v, ",ok=", ok)
}

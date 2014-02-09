package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	str := `{
		"id": 1985,
	}`
	var o interface {}
	err := json.Unmarshal([]byte(str), &o)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(o)
}

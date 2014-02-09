package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	var str = `{
		"id": "1945",
		"users":[
			{"name":"winlin", "age":28, "address":"beijing"},
			{"name":"stone", "age":29, "address":"china"}
		],
		"servers":[
			{"ip":"192.168.1.2", "desc":"svn"},
			{"ip":"192.168.1.5", "desc":"file"},
			{"ip":"192.168.1.3", "desc":"db"}
		]
	}`

	var m interface {}

	err := 	json.Unmarshal([]byte(str), &m)
	if err != nil {
		fmt.Println("error: ", err)
	}
	fmt.Printf("%+v\n", m)

	type JsonAny map[string] interface {}
	var mm JsonAny
	var ok bool
	if mm,ok = m.(JsonAny); true {
		fmt.Printf("ok=%v, id=%v\n", ok, mm["id"])
	}
	if mm,ok = m.(map[string]interface {}); true {
		fmt.Printf("ok=%v, id=%v\n", ok, mm["id"])
	}
	// panic: interface conversion: interface is map[string]interface {}, not main.JsonAny
	//mm = m.(JsonAny)

	b,err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
}

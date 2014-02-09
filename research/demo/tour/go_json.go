package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type User struct {
	// the field must be exported, or the json will never marshal/unmarshal it.
	Name string
	Age int
	Address string
}
type Server struct {
	Ip string
	Desc string
}
type Data struct {
	Id string
	Users []User
	Servers []Server
}

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

	var m Data

	err := 	json.Unmarshal([]byte(str), &m)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", m)

	b,err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
}

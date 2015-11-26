package main

import (
	"flag"
	"net/http"

	myhttp "github.com/Alienero/IamServer/http"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush()
	if err := flag.Set("logtostderr", "true"); err != nil {
		panic(err)
	}
	r := NewSrsServer()
	r.PrintInfo()
	// init http server
	if err := myhttp.InitHTTP(); err != nil {
		panic(err)
	}
	go func() {
		if err := http.ListenAndServe(":9090", nil); err != nil {
			panic(err)
		}
	}()
	r.Serve()
}

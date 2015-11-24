package main

import (
	"flag"

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
	r.Serve()
}

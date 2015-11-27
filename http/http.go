package http

import (
	"net/http"

	"github.com/Alienero/IamServer/source"

	"github.com/golang/glog"
)

func InitHTTP() error {
	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		glog.Info("http: get an request.", r.RequestURI, r.Method)
		if r.Method != "GET" {
			return
		}
		// get live source.
		// TODO: should map source's http request and source key.
		key := "/live/123" // for test.
		consumer, err := source.NewConsumer(key)
		if err != nil {
			glog.Info("<<<<<<<<<< can not get source >>>>>>>>>>>>>", err)
			return
		}
		defer consumer.Close()

		glog.Info("<<<<<<<<<<<<<<<<<<<<<<<<<got source>>>>>>>>>>>>>>>>>>>>>>>>>")
		// set flv live stream http head.
		// TODO: let browser not cache sources.
		w.Header().Add("Content-Type", "video/x-flv")
		if err := consumer.Live(w); err != nil {
			glog.Info("Live get an client error:", err)
		}
	})
	http.Handle("/", http.FileServer(http.Dir(".")))
	return nil
}

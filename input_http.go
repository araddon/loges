package loges

import (
	u "github.com/araddon/gou"
	"github.com/bmizerany/pat"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func MakeCustomHandler(msgsOut chan *LineEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		qs := r.URL.Query()
		stream := qs.Get("stream")
		if len(stream) < 1 {
			stream = qs.Get(":stream")
			if len(stream) < 1 {
				io.WriteString(w, "Requires a 'stream' qs param ")
				return
			} else {
				qs.Del(":stream")
			}
		}
		var data []byte
		var err error
		if r.Body != nil {
			data, err = ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				u.Log(u.ERROR, err.Error())
				io.WriteString(w, "Requires valid monit parse")
				return
			}
		} else {
			data = []byte(qs.Encode())
		}
		u.Debug(stream, string(data))
		//u.Error("Not implemented")
		msgsOut <- &LineEvent{Data: data, Source: stream}
	}
}

func MakeMonitHandler(msgsOut chan *LineEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		qs := r.URL.Query()
		stream := qs.Get(":stream")

		if len(stream) < 1 {
			io.WriteString(w, "Requires a 'stream' qs param")
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		//u.Debug(string(data))
		defer r.Body.Close()
		if err != nil {
			u.Log(u.ERROR, err.Error())
			io.WriteString(w, "Requires valid monit parse")
			return
		}
		nv, timeNs := MonitParse(data)
		u.Debug(nv, timeNs)
		msgsOut <- &LineEvent{Data: []byte(nv.Encode()), Source: "monit"}
	}
}

// We are starting an http ->  collector
//  accepts monit data
//
func HttpRun(httpPort string, msgsOut chan *LineEvent) {

	m := pat.New()
	monitHandler := http.HandlerFunc(MakeMonitHandler(msgsOut))
	httpCustomHandler := http.HandlerFunc(MakeCustomHandler(msgsOut))
	m.Post("/monit/:stream", monitHandler)
	m.Get("/", httpCustomHandler)
	m.Post("/", httpCustomHandler)

	http.Handle("/", m)
	httpAddr := ":" + httpPort
	u.Logf(u.WARN, "Starting http on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	u.Log(u.INFO, "All services have shut down, quitting")

}

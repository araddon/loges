package loges

import (
	u "github.com/araddon/gou"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"time"
)

// read the message channel and send to elastic search
// Uses the background Bulk Indexor, which has the **Possibility** of losing
// data if app panic/quits (but is much faster than non-bulk)
func ToElasticSearch(msgChan chan *LineEvent, esType, esHost, ttl string) {
	// set elasticsearch host which is a global
	u.Warnf("Starting elasticsearch on %s", esHost)
	api.Domain = esHost
	done := make(chan bool)
	indexor := core.NewBulkIndexorErrors(20, 120)
	indexor.Run(done)

	errorCt := 0 // use sync.atomic or something if you need
	timer := time.NewTicker(time.Minute * 1)
	go func() {
		for {
			select {
			case _ = <-timer.C:
				if errorCt < 2 {
					errorCt = 0
				} else {
					panic("Too many errors in ES")
				}
			case _ = <-done:
				return
			}
		}
	}()
	for errBuf := range indexor.ErrorChannel {
		errorCt++
		u.Error(errBuf.Err)
		// log to disk?  db?   ????  Panic
	}

	for in := range msgChan {
		msg := formatter(in)
		if msg != nil {
			if err := core.IndexBulkTtl(msg.Index(), esType, msg.Id(), ttl, nil, msg); err != nil {
				u.Error("%v", err)
			}
		}
	}
}

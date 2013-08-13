package loges

import (
	"bytes"
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
	indexor.BulkSendor = func(buf *bytes.Buffer) error {
		//u.Debug(string(buf.Bytes()))
		return core.BulkSend(buf)
	}
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
	go func() {
		for errBuf := range indexor.ErrorChannel {
			errorCt++
			u.Error(errBuf.Err)
			// log to disk?  db?   ????  Panic
		}
	}()

	u.Debug("Starting MsgChan to ES ", len(msgChan))
	// TODO, refactor this and stdout one into a "Router"
	for in := range msgChan {
		for _, transform := range transforms {
			if msg := transform(in); msg != nil {
				if err := indexor.Index(msg.Index(), esType, msg.Id(), ttl, nil, msg); err != nil {
					u.Error("%v", err)
				}
			} else {
				break
			}
		}
	}
}

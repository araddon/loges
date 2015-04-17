package loges

import (
	"bytes"
	"strings"
	"time"

	u "github.com/araddon/gou"
	elastigo "github.com/mattbaird/elastigo/lib"
)

// read the message channel and send to elastic search
// Uses the background Bulk Indexor, which has the **Possibility** of losing
// data if app panic/quits (but is much faster than non-bulk)
func ToElasticSearch(msgChan chan *LineEvent, esType, esHost, ttl string, sendMetrics bool) {
	// set elasticsearch host which is a global
	u.Warnf("Starting elasticsearch on %s", esHost)
	elastigoConn := elastigo.NewConn()
	// The old standard for host was including :9200
	esHost = strings.Replace(esHost, ":9200", "", -1)
	elastigoConn.SetHosts([]string{esHost})
	indexer := elastigoConn.NewBulkIndexerErrors(10, 120)
	//indexer := elastigo.NewBulkIndexerErrors(20, 120)
	indexer.Sender = func(buf *bytes.Buffer) error {
		//u.Debug(string(buf.Bytes()))
		u.Infof("es writing: %d bytes", buf.Len())
		return indexer.Send(buf)
	}
	indexer.Start()

	errorCt := 0 // use sync.atomic or something if you need
	timer := time.NewTicker(time.Minute * 2)
	go func() {
		for {
			select {
			case _ = <-timer.C:
				u.Infof("errorCt: %d", errorCt)
				if errorCt < 5 {
					errorCt = 0
				} else {
					panic("Too many errors in ES")
				}
			}
		}
	}()
	go func() {
		for errBuf := range indexer.ErrorChannel {
			errorCt++
			u.Error(errBuf.Err)
			// log to disk?  db?   ????  Panic
		}
	}()

	//u.Debug("Starting MsgChan to ES ", len(msgChan))
	// TODO, refactor this and stdout one into a "Router"
	for in := range msgChan {
		for _, transform := range transforms {
			if msg := transform(in); msg != nil {
				if !sendMetrics && (in.DataType == "METRIC" || in.DataType == "METR") {
					continue
				}
				if err := indexer.Index(msg.Index(), esType, msg.Id(), ttl, nil, msg, false); err != nil {
					u.Error("%v", err)
				}
			} else {
				//These are ok, just means its not destined for ElasticSearch
				//u.Debugf("bad es? %v", in)
			}
		}
	}
}

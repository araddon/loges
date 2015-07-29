package loges

import (
	"bytes"
	"os"
	"strings"
	"time"

	u "github.com/araddon/gou"
	elastigo "github.com/mattbaird/elastigo/lib"
)

// read the message channel and send to elastic search
// Uses the background Bulk Indexor, which has the **Possibility** of losing
// data if app panic/quits (but is much faster than non-bulk)
//  @exitIfNoMsgs :  Should we panic if we don't see messages after this duration?
func ToElasticSearch(msgChan chan *LineEvent, esType, esHost, ttl string,
	exitIfNoMsgs time.Duration, sendMetrics bool) {

	checkForMsgs := false
	if exitIfNoMsgs.Seconds() > 0 {
		checkForMsgs = true
	}

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
	lastMsgTime := time.Now()
	msgCt := 0
	go func() {
		for {
			select {
			case _ = <-timer.C:
				u.Infof("errorCt: %d", errorCt)
				if errorCt < 5 {
					// We reset errors back to 0, if we didn't climb too high
					// so that they don't continually grow
					errorCt = 0
				} else {
					// If we can't write to ES, really not much else we can do
					// here other than panic, sending possibly an alert to
					// our monit?
					u.Errorf("We have too many errors, exiting: %v", errorCt)
					os.Exit(1)
				}

				if checkForMsgs {
					if time.Now().After(lastMsgTime.Add(exitIfNoMsgs)) {
						u.Errorf("We have not seen a message since %v secs ago, exiting: msgs:%v lastmsg:%v",
							time.Now().Sub(lastMsgTime), msgCt, lastMsgTime)
						os.Exit(1)
					} else {
						u.Infof("just processed %v msgs %v", msgCt, lastMsgTime)
					}
				}
			}
		}
	}()
	go func() {
		for errBuf := range indexer.ErrorChannel {
			errorCt++
			u.Errorf("ES Indexer Err: %v\n", errBuf.Err)
			// log to disk?  db?   ????  Panic
		}
	}()

	//u.Debug("Starting MsgChan to ES ", len(msgChan))
	// TODO, refactor this and stdout one into a "Router"
	for in := range msgChan {
		for _, transform := range transforms {
			if msg := transform(in); msg != nil {
				if in.DataType == "METRIC" || in.DataType == "METR" {
					if !sendMetrics {
						continue
					}
				} else {
					lastMsgTime = time.Now()
					msgCt += 1
				}

				if err := indexer.Index(msg.Index(), esType, msg.Id(), ttl, nil, msg, false); err != nil {
					u.Errorf("Index(ing) error: %v\n", err)

					//Increment WriteErrs log and requeue message
					in.WriteErrs += 1
					msgChan <- in
				}
			} else {
				//These are ok, just means its not destined for ElasticSearch
				//u.Debugf("bad es? %v", in)
			}
		}
	}
}

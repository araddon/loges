package main

import (
	"flag"
	"github.com/araddon/loges"
	"os"
	"strings"
	"time"
)

var (
	msgChan     = make(chan *loges.LineEvent, 1000)
	hostname    string
	logType     string
	loc                = time.UTC
	currentHost string = "Unknown"
)

func init() {
	flag.StringVar(&hostname, "eshost", "localhost:9200", "host string for the Elasticsearch Server")
	flag.StringVar(&logType, "logtype", "datatype", "Type of data for elasticsearch index")
	if host, err := os.Hostname(); err == nil {
		currentHost = host
	}
}

func main() {
	flag.Parse()

	// 1.  Start Output
	// update the logstash index occasionally
	go loges.UpdateLogstashIndex()
	// start an elasticsearch writer worker, for sending to elasticsearch
	go loges.ToElasticSearch(msgChan, "golog", cleanEsHost(hostname))

	// 2.  Format/Filter
	// create our custom formatter for parsing/filtering,/manipulating line entries
	loges.FormatterSet(CustomFormatter(200))

	// 3.  Input:  Start our Input and block
	loges.StdinPruducer(msgChan)
}

func CustomFormatter(every uint64) loges.LineFormatter {
	return func(d *loges.LineEvent) *loges.Event {
		line := string(d.Data)
		if strings.Contains(line, "DELETEME") {
			return nil
		}
		parts := strings.SplitN(line, " ", 4)
		if len(parts) < 2 {
			return nil
		}
		// custom formatting
		if _, err := time.Parse("2006/01/02 15:04:05.000000", parts[0]); err == nil {
			evt := loges.NewTsEvent("mytype", d.Source, parts[2], time.Now().In(loc))
			evt.Fields = make(map[string]interface{})
			evt.Fields["host"] = currentHost
			return evt
		}
		// if we return nil, it is filtered out
		return nil
	}
}

func cleanEsHost(oldHost string) string {
	// It is possible that they sent a list of posts with :9200
	hosts := strings.Split(oldHost, ",")
	if len(hosts) > 0 {
		parts := strings.Split(hosts[0], ":")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return oldHost
}

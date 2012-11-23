package main

import (
	"flag"
	"github.com/araddon/loges"
	"math"
)

var (
	msgChan      = make(chan *loges.LineEvent, 1000)
	hostname     string
	logConfig    string
	logLevel     string
	logType      string
	esIndex      string
	format       string
	kafkaHost    string
	topic        string
	partitionstr string
	offset       uint64
	maxSize      uint
	maxMsgCt     uint64
)

func init() {
	flag.StringVar(&hostname, "hostname", "lio26", "host:port string for the kafka server")
	flag.StringVar(&logConfig, "logconfig", "timber.xml", "file for logging config")
	flag.StringVar(&logLevel, "loglevel", "INFO", "loglevel [NONE,DEBUG,INFO,WARNING,ERROR]")
	flag.StringVar(&format, "source", "fluentd", "Format [fluentd,kafka]")
	flag.StringVar(&logType, "logtype", "datatype", "Type of data for elasticsearch index")
	// kafka config info
	flag.StringVar(&kafkaHost, "hostname", "localhost:9092", "host:port string for the kafka server")
	flag.StringVar(&topic, "topic", "test", "topic to publish to")
	flag.StringVar(&partitionstr, "partitions", "0", "partitions to publish to:  comma delimited")
	flag.Uint64Var(&offset, "offset", 0, "offset to start consuming from")
	flag.UintVar(&maxSize, "maxsize", 1048576, "max size in bytes to consume a message set")
	flag.Uint64Var(&maxMsgCt, "msgct", math.MaxUint64, "max number of messages to read")
}

func main() {
	flag.Parse()
	loges.TimberSetLogging("[%D %T] %s %L %M", logLevel)
	// update the index occasionally
	go loges.UpdateLogstashIndex()
	// start an elasticsearch writer worker, for sending to elasticsearch
	go loges.ToElasticSearch(msgChan, "golog", hostname)
	switch format {
	case "fluentd":
		loges.FormatterSet(loges.FluentdFormatter(logType, nil))
		go loges.StdinPruducer(msgChan)
	case "kafka":
		loges.FormatterSet(loges.KafkaFormatter)
		//partitionstr, topic, kafkaHost string, offset, maxMsgCt uint64, maxSize uint
		go loges.RunKafkaConsumer(msgChan, partitionstr, topic, kafkaHost, offset, maxMsgCt, maxSize)
	}

	done := make(chan byte)
	<-done
}

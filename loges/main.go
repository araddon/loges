package main

import (
	"flag"
	"fmt"
	u "github.com/araddon/gou"
	"github.com/araddon/loges"
	"github.com/araddon/loges/kafka"
	tail "github.com/fw42/go-tail"
	"math"
	"os"
	"strings"
)

var (
	msgChan      = make(chan *loges.LineEvent, 1000)
	esHostName   string
	logConfig    string
	logLevel     string
	logType      string
	esIndex      string
	source       string
	filter       string
	kafkaHost    string
	topic        string
	output       string
	partitionstr string
	offset       uint64
	maxSize      uint
	maxMsgCt     uint64
	colorize     bool
)

func Usage() {
	usage := `Usage of %s:

For tail, pass arguments at command line:
	
	# take two files, output to stdout, using gofile formatter
	loges --filter=gofiles \
		/path/to/myfile.log \
		/path/to/myfile2.log 

	# tail this api.log file to elasticsarch
	loges --filter=gofiles --out=elasticsearch \
		--eshost=192.168.1.13 --loglevel=NONE \
		/mnt/log/api.log

	# read stdin and send to stdout
	myapp | loges --source=stdin --filter=stdfiles 
`
	fmt.Fprintf(os.Stderr, usage, os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.Usage = Usage
	flag.StringVar(&esHostName, "eshost", "localhost", "host (no port) string for the elasticsearch server")
	flag.StringVar(&logLevel, "loglevel", "DEBUG", "loglevel [NONE,DEBUG,INFO,WARNING,ERROR]")
	flag.StringVar(&source, "source", "tail", "Format [stdin,kafka,tail]")
	flag.StringVar(&filter, "filter", "fluentd", "Filter to apply [stdfiles,fluentd]")
	flag.StringVar(&output, "out", "stdout", "Output destiation [elasticsearch, stdout]")
	flag.StringVar(&logType, "logtype", "stdfiles", "Type of data for elasticsearch index")
	flag.BoolVar(&colorize, "colorize", true, "Colorize Stdout?")
	// kafka config info
	flag.StringVar(&kafkaHost, "kafkahost", "localhost:9092", "host:port string for the kafka server")
	flag.StringVar(&topic, "topic", "test", "topic to publish to")
	flag.StringVar(&partitionstr, "partitions", "0", "partitions to publish to:  comma delimited")
	flag.Uint64Var(&offset, "offset", 0, "offset to start consuming from")
	flag.UintVar(&maxSize, "maxsize", 1048576, "max size in bytes to consume a message set")
	flag.Uint64Var(&maxMsgCt, "msgct", math.MaxUint64, "max number of messages to read")
}

func main() {
	flag.Parse()
	done := make(chan bool)
	u.SetupLogging(logLevel)
	u.SetColorIfTerminal()
	esHostName = cleanEsHost(esHostName)
	u.Debugf("Connecting to ES:  %s", esHostName)

	// Setup output first, to ensure its ready when Source starts
	switch output {
	case "elasticsearch":
		// update the Logstash date for the index occasionally
		go loges.UpdateLogstashIndex()
		// start an elasticsearch bulk worker, for sending to elasticsearch
		go loges.ToElasticSearch(msgChan, "golog", esHostName)
	case "stdout":
		u.Error("setting output to stdout")
		go loges.ToStdout(msgChan, colorize)
	default:
		println("No output set, required")
		Usage()
		os.Exit(1)
	}

	// now set up the formatter/filters
	//for _, filter := range strings.Split(filters, ",") {}
	switch filter {
	case "stdfiles":
		loges.FormatterSet(loges.FileFormatter(logType, nil))
	case "fluentd":
		loges.FormatterSet(loges.FluentdFormatter(logType, nil))
	case "kafka":
		loges.FormatterSet(kafka.KafkaFormatter)
	}

	switch source {
	case "tail":
		config := tail.Config{Follow: true, ReOpen: true}
		for _, filename := range flag.Args() {
			go loges.TailFile(filename, config, done, msgChan)
		}
	case "kafka":
		//partitionstr, topic, kafkaHost string, offset, maxMsgCt uint64, maxSize uint
		go kafka.RunKafkaConsumer(msgChan, partitionstr, topic, kafkaHost, offset, maxMsgCt, maxSize)
	case "stdin":
		go loges.StdinPruducer(msgChan)
	default: // "stdin"
		println("No input set, required")
		Usage()
		os.Exit(1)
	}

	<-done
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

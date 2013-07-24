package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/tail"
	u "github.com/araddon/gou"
	"github.com/araddon/loges"
	"github.com/araddon/loges/kafka"
	"log"
	"math"
	"os"
	"strings"
)

var (
	msgChan        = make(chan *loges.LineEvent, 1000)
	metricsChan    = make(chan *loges.LineEvent, 1000)
	esHostName     string
	logConfig      string
	logLevel       string
	logType        string
	esIndex        string
	source         string
	graphiteHost   string
	graphitePrefix string
	filters        string
	httpPort       string
	kafkaHost      string
	topic          string
	output         string
	metricsOut     string
	partitionstr   string
	ttl            string
	offset         uint64
	maxSize        uint
	maxMsgCt       uint64
	colorize       bool
	_              = log.Ldate
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

	# read stdin and send to stdout colorized
	myapp | loges 
`
	fmt.Fprintf(os.Stderr, usage, os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.Usage = Usage
	flag.StringVar(&esHostName, "eshost", "localhost", "host (no port) string for the elasticsearch server")
	flag.StringVar(&logLevel, "loglevel", "DEBUG", "loglevel [NONE,DEBUG,INFO,WARNING,ERROR]")
	flag.StringVar(&source, "source", "tail", "Comma Delimited Data Sources for log file/data [stdin,kafka,tail,http] (http=monit)")
	flag.StringVar(&filters, "filter", "stdfiles", "Filter to apply [stdfiles,fluentd]")
	flag.StringVar(&output, "out", "stdout", "Output destination [elasticsearch, stdout]")
	flag.StringVar(&metricsOut, "metrics", "", "Output for metrics [librato,graphite,]")
	flag.StringVar(&logType, "logtype", "stdfiles", "Type of data for elasticsearch index")
	flag.BoolVar(&colorize, "colorize", false, "Colorize Stdout?")
	flag.StringVar(&httpPort, "port", "8398", "Port number for http service")
	flag.StringVar(&graphiteHost, "graphite", "carbon.hostedgraphite.com:2003", "host for graphite")
	flag.StringVar(&graphitePrefix, "gprefix", "", "graphite prefix")
	// kafka config info
	flag.StringVar(&kafkaHost, "kafkahost", "localhost:9092", "host:port string for the kafka server")
	flag.StringVar(&topic, "topic", "test", "topic to publish to")
	flag.StringVar(&partitionstr, "partitions", "0", "partitions to publish to:  comma delimited")
	flag.Uint64Var(&offset, "offset", 0, "offset to start consuming from")
	flag.UintVar(&maxSize, "maxsize", 1048576, "max size in bytes to consume a message set")
	flag.Uint64Var(&maxMsgCt, "msgct", math.MaxUint64, "max number of messages to read")
	flag.StringVar(&ttl, "ttl", "30d", "Elasticsearch TTL ")
}

func main() {
	flag.Parse()
	u.SetupLogging(logLevel)
	u.SetColorIfTerminal() // this doesn't work if reading stdin
	if colorize {
		u.SetColorOutput()
	}

	done := make(chan bool)
	esHostName = cleanEsHost(esHostName)
	// if we have note specified tail files, then assume stdin
	if len(flag.Args()) == 0 && source == "tail" {
		source = "stdin"
	}

	u.Debugf("LOGES: filters=%s  es=%s argct=:%d source=%v ll=%s",
		filters, esHostName, len(flag.Args()), source, logLevel)

	// Setup output first, to ensure its ready when Source starts
	// TODO:  suuport multiple outputs?
	switch output {
	case "elasticsearch":
		// update the Logstash date for the index occasionally
		go loges.UpdateLogstashIndex()
		// start an elasticsearch bulk worker, for sending to elasticsearch
		go loges.ToElasticSearch(msgChan, "golog", esHostName, ttl)
	case "stdout":
		u.Debug("setting output to stdout ", colorize)
		go loges.ToStdout(msgChan, colorize)
	default:
		Usage()
		os.Exit(1)
	}

	// TODO:  implement metrics out
	for _, metOut := range strings.Split(metricsOut, ",") {
		switch metOut {
		case "librato":
			//
		case "graphite":
			u.Info("Registering Graphite Transform")
			loges.TransformRegister(loges.GraphiteTransform(graphiteHost, graphitePrefix))
		}
	}

	// now set up the transforms/filters
	for _, filter := range strings.Split(filters, ",") {
		switch filter {
		case "stdfiles":
			loges.TransformRegister(loges.FileFormatter(logType, nil))
		case "fluentd":
			loges.TransformRegister(loges.FluentdFormatter(logType, nil))
		case "kafka":
			loges.TransformRegister(kafka.KafkaFormatter)
		}
	}

	for _, sourceInput := range strings.Split(source, ",") {
		u.Warnf("source = %v", sourceInput)
		switch sourceInput {
		case "tail":
			for _, filename := range flag.Args() {
				tailDone := make(chan bool)
				go loges.TailFile(filename, tail.Config{Follow: true, ReOpen: true}, tailDone, msgChan)
			}
		case "http":
			go loges.HttpRun(httpPort, msgChan)
		case "kafka":
			go kafka.RunKafkaConsumer(msgChan, partitionstr, topic, kafkaHost, offset, maxMsgCt, maxSize)
		case "stdin":
			go loges.StdinPruducer(msgChan)
		default:
			u.Error(sourceInput)
			println("No input set, required")
			Usage()
			os.Exit(1)
		}
	}
	u.Warn("end of main startup, until done")
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

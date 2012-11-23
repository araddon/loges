Go Logtash Writer
---------------------------------

A go library for writing to Elasticsearch using Logstash format. Reads Fluentd format from stdin and kafka.

Why?  
---------
I wanted to send data to Elasticsearch for viewing in http://kibana.org/ from within a go app for debugging purposes.  But also wanted to send some log files using http://fluentd as well and choose go instead of native ruby support provided by Fluentd.


Features
-----------------

  * Writes to elasticsearch using http://logstash.net/ format to allow usage of http://kibana.org/.   
  * Has readers for Stdin
  * Formatter for parsing fluentd format http://fluentd.org/
  * reads from Kafka 
  * custom formatters allowed (see below)


Create Custom Formatter (see /example/main.go)::
	
	// custom formatter
	func CustomFormatter(every uint64) loges.LineFormatter {
		return func(e *loges.LineEvent) *loges.Event {
			// custom formatting
			return nil
		}
	}




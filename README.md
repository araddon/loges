Go Logtash Writer
---------------------------------

Writes to elasticsearch using http://logstash.net/ format to allow usage of http://kibana.org/.   

Has readers for Stdin that assumes http://fluentd.org/ format. 

Also reads from Kafka to send.


Create Custom Formatter (see /example/main.go)::
	
	// custom formatter
	func CustomFormatter(every uint64) loges.LineFormatter {
		return func(e *loges.LineEvent) *loges.Event {
			// custom formatting
			return nil
		}
	}




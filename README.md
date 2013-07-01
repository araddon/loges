Logging Utility & Daemon
---------------------------------

Read log data from Inputs [Tail Files, Kafka, Stdin], perform transforms 
[Combine lines(multi-line-errors)] and output to [ElasticSearch, Stdout]
Recognize lines with metrics and send to Monitoring [Graphite, Librato, ..]


![Drawing](https://docs.google.com/drawings/d/1nGVabfy3PB0Zq-gsghKRkGU3eGz4zrcmpIrB0e2cs9M/pub?w=695&h=401)

Why?  
---------
I wanted to send data to Elasticsearch for viewing in http://kibana.org/ and also wanted
to read from kafka.  http://fluentd does this well but I choose go instead of native ruby support provided by Fluentd.

There are probably better tools out there for this but putting together the 
specific combination i wanted (LogStash format in Elasticsearch, Tail files, Read Kafka), see alternates below.


Features
-----------------

* **Inputs**
  * Stdio (Can read FluentD format)
  * Kafka
  * Tail Logs (multiple files)
* **Transforms**:
   * Logstash http://logstash.net/ 
   * Colorizor for console
   * Concat into single line when needed (say error stack trace)
   * custom plugins
* **Outputs**
   * Stdout
   * Elasticsearch

Alternates
-----------------

* Go https://github.com/onemorecloud/dendrite
* Go https://github.com/ryandotsmith/l2met
* Go http://blog.mozilla.org/services/2013/04/30/introducing-heka/
* Ruby, http://fluentd.org/
* JVM, http://logstash.org 

Extending
----------------------

Create Custom Formatter (see /example/main.go)::
	
	// custom formatter
	func CustomFormatter(every uint64) loges.LineFormatter {
		return func(e *loges.LineEvent) *loges.Event {
			// custom formatting
			return nil
		}
	}




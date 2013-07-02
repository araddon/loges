Logging Utility & Daemon
---------------------------------

Read log data from Inputs [Tail Files, Kafka, Stdin], perform transforms 
[Combine lines(multi-line-errors)] and output to [ElasticSearch, Stdout]
Recognize lines with metrics and send to Monitoring [Graphite, Librato, ..]


![Drawing](https://docs.google.com/drawings/d/1nGVabfy3PB0Zq-gsghKRkGU3eGz4zrcmpIrB0e2cs9M/pub?w=695&h=401)

Why?  
---------
We had 2 needs:  1) to send data to Elasticsearch for viewing in http://kibana.org/ 
and 2) if possible, unify the Logging/Metrics systems data-collection-forwarding.  

There are probably better tools out there for this but putting together the 
specific combination of: (LogStash format in Elasticsearch, Tail files, 
Read Kafka, Metrics read from log files) didn't happen, see alternates below.


Features
-----------------

* **Inputs**
  * Stdin 
  * Kafka
  * Tail Logs (multiple files)
  * Monit (metrics vi http)
  * Http  (custom)
* **Transforms**:
   * Logstash http://logstash.net/ 
   * Colorizor for console
   * Concat into single line when needed (say error stack trace)
   * custom plugins
   * Seperate Metrics Log Lines from regular log lines
* **Log Line Outputs**
   * Stdout (optional colorized)
   * Elasticsearch
* **Metric Outputs**
   * Graphite

Alternates
-----------------

* Go https://github.com/onemorecloud/dendrite
* Go https://github.com/ryandotsmith/l2met
* Go http://blog.mozilla.org/services/2013/04/30/introducing-heka/
* Ruby, http://fluentd.org/
* JVM, http://logstash.org 
* Go https://github.com/cloudfoundry/loggregator

Usage
----------------------

```sh
loges --source=monit,tail --metrics=graphite \
   /path/to/my/file \
   /path/to/another
```


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




Go Logging Utilitity
---------------------------------

Read log type data from Kafka, Stdin, or Files and then Perform transforms on it
(Colorize, Transform to Logstash format) and then output to ElasticSearch or Stdout


Why?  
---------
I wanted to send data to Elasticsearch for viewing in http://kibana.org/ from within a go app for debugging purposes.  But also wanted to send some log files using http://fluentd as well and choose go instead of native ruby support provided by Fluentd.

There are probably better tools out there for this but putting together the 
specific combination i wanted (LogStash format in Elasticsearch, Tail files, Read Kafka), see alternates below.


Features
-----------------

* **Inputs**
  * Stdio (Can read FluentD format)
  * Kafka
  * Tail Logs (multiple files)
* **Formatters**:
   * Logstash http://logstash.net/ 
   * Colorizor for console
   * custom plugin
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




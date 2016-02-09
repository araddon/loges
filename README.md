Logging Utility & Daemon
---------------------------------

Read log data from Inputs [Tail Files, Stdin, Monit], perform transforms 
[Combine lines(multi-line-errors)] and output to [ElasticSearch, Stdout]
Recognize lines with metrics and send to Monitoring [Graphite, InfluxDB, ..]


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
  * Tail Logs (multiple files)
  * Monit (metrics via http)
  * Http  (custom)
* **Transforms**:
   * Logstash http://logstash.net/ 
   * Colorizer for console
   * Concat into single line when needed (e.g. error stack trace)
   * Separate Metrics Log Lines from regular log lines
   * Custom plugins
* **Log Line Outputs**
   * Stdout (optional colorized)
   * Elasticsearch
* **Metric Outputs**
   * Graphite

Alternatives
-----------------

* Go https://github.com/gliderlabs/logspout
* Go https://github.com/onemorecloud/dendrite
* Go https://github.com/ryandotsmith/l2met
* Go http://blog.mozilla.org/services/2013/04/30/introducing-heka/
* Ruby, http://fluentd.org/
* JVM, http://logstash.org 
* Go https://github.com/cloudfoundry/loggregator

Usage
----------------------

```sh
loges --source=monit,tail --filter=stdfiles --out=elasticsearch --metrics=graphite \
   /path/to/my/file \
   /path/to/another/file

```



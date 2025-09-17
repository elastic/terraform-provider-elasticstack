Helper data source which can be used to create the configuration for a grok processor. This processor extracts structured fields out of a single text field within a document. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/grok-processor.html

This processor comes packaged with many [reusable patterns](https://github.com/elastic/elasticsearch/blob/master/libs/grok/src/main/resources/patterns).

If you need help building patterns to match your logs, you will find the [Grok Debugger](https://www.elastic.co/guide/en/kibana/master/xpack-grokdebugger.html) tool quite useful! [The Grok Constructor](https://grokconstructor.appspot.com/) is also a useful tool.
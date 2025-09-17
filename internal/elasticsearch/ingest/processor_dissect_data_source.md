Helper data source which can be used to create the configuration for a dissect processor. This processor extracts structured fields out of a single text field within a document. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/dissect-processor.html#dissect-processor

Similar to the Grok Processor, dissect also extracts structured fields out of a single text field within a document. However unlike the Grok Processor, dissect does not use Regular Expressions. This allows dissect’s syntax to be simple and for some cases faster than the Grok Processor.

Dissect matches a single text field against a defined pattern.
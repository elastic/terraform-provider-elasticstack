Helper data source which can be used to create the configuration for a bytes processor. The processor converts a human readable byte value (e.g. 1kb) to its value in bytes (e.g. 1024). See the [bytes processor documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/bytes-processor.html) for more details.

If the field is an array of strings, all members of the array will be converted.

Supported human readable units are "b", "kb", "mb", "gb", "tb", "pb" case insensitive. An error will occur if the field is not a supported format or resultant value exceeds 2^63.
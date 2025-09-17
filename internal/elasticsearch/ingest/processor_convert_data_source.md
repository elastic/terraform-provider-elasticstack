Helper data source which can be used to create the configuration for a convert processor. This processor converts a field in the currently ingested document to a different type, such as converting a string to an integer. See the [convert processor documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/convert-processor.html) for more details.

The supported types include: 
- `integer`
- `long`
- `float`
- `double`
- `string`
- `boolean`
- `ip`
- `auto`

Specifying `boolean` will set the field to true if its string value is equal to true (ignoring case), to false if its string value is equal to false (ignoring case), or it will throw an exception otherwise.

Specifying `ip` will set the target field to the value of `field` if it contains a valid IPv4 or IPv6 address that can be indexed into an IP field type.

Specifying `auto` will attempt to convert the string-valued `field` into the closest non-string, non-IP type. For example, a field whose value is "true" will be converted to its respective boolean type: true. Do note that float takes precedence of double in auto. A value of "242.15" will "automatically" be converted to 242.15 of type `float`. If a provided field cannot be appropriately converted, the processor will still process successfully and leave the field value as-is. In such a case, `target_field` will be updated with the unconverted field value.
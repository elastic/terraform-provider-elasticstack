Helper data source which can be used to create the configuration for a foreach processor. This processor runs an ingest processor on each element of an array or object. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/foreach-processor.html

All ingest processors can run on array or object elements. However, if the number of elements is unknown, it can be cumbersome to process each one in the same way.

The `foreach` processor lets you specify a `field` containing array or object values and a `processor` to run on each element in the field.

### Access keys and values

When iterating through an array or object, the foreach processor stores the current element’s value in the `_ingest._value` ingest metadata field. `_ingest._value` contains the entire element value, including any child fields. You can access child field values using dot notation on the `_ingest._value` field.

When iterating through an object, the foreach processor also stores the current element’s key as a string in `_ingest._key`.

You can access and change `_ingest._key` and `_ingest._value` in the processor.
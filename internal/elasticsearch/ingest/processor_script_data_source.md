Helper data source which can be used to create the configuration for a script processor. This processor runs an inline or stored script on incoming documents. See the [script processor documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/script-processor.html) for more details.

The script processor uses the script cache to avoid recompiling the script for each incoming document. To improve performance, ensure the script cache is properly sized before using a script processor in production.

### Access source fields

The script processor parses each incoming document’s JSON source fields into a set of maps, lists, and primitives. To access these fields with a Painless script, use the map access operator: `ctx['my-field']`. You can also use the shorthand `ctx.<my-field>` syntax.

### Access metadata fields

You can also use a script processor to access metadata fields.
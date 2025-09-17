Helper data source which can be used to create the configuration for a geoip processor. The geoip processor adds information about the geographical location of an IPv4 or IPv6 address. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/geoip-processor.html

By default, the processor uses the GeoLite2 City, GeoLite2 Country, and GeoLite2 ASN GeoIP2 databases from MaxMind, shared under the CC BY-SA 4.0 license. Elasticsearch automatically downloads updates for these databases from the Elastic GeoIP endpoint: https://geoip.elastic.co/v1/database. To get download statistics for these updates, use the GeoIP stats API.

If your cluster can’t connect to the Elastic GeoIP endpoint or you want to manage your own updates, [see Manage your own GeoIP2 database updates](https://www.elastic.co/guide/en/elasticsearch/reference/current/geoip-processor.html#manage-geoip-database-updates).

If Elasticsearch can’t connect to the endpoint for 30 days all updated databases will become invalid. Elasticsearch will stop enriching documents with geoip data and will add `tags: ["_geoip_expired_database"]` field instead.

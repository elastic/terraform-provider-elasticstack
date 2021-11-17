# Using this provider with Elastic Cloud

This example depicts how to spin up an [Elastic Cloud](https://cloud.elastic.co) deployment containing an Elasticsearch and a Kibana cluster. 
Then, using the stack provider, we connect to the newly created deployment and configure an ingest user and index template on that cluster.
Everything is applied in a single terraform apply.

## Running the example

To run the example, follow these steps:

1. Make sure you are signed up to Elastic Cloud, and that you have an API key.
2. Apply your API key to the provider.tf file.
3. Run `terrafrom init` to initialize your Terraform CLI.
4. Run `terraform apply` to see how it works.

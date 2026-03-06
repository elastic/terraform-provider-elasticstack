# ML Job State Resource

Manages the state of an Elasticsearch Machine Learning (ML) job, allowing you to open or close ML jobs.

This resource uses the following Elasticsearch APIs:
- [Open ML Job API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-open-job.html)
- [Close ML Job API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-close-job.html)
- [Get ML Job Stats API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-job-stats.html)

## Important Notes

- This resource manages the **state** of an existing ML job, not the job configuration itself.
- The ML job must already exist before using this resource.
- Opening a job allows it to receive and process data.
- Closing a job stops data processing and frees up resources.
- Jobs can be opened and closed multiple times throughout their lifecycle.
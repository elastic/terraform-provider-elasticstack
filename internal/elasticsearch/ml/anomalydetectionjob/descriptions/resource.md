Creates and manages Machine Learning anomaly detection jobs.

See the [ML Job API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-job.html) for more details.

## Migration note: `timeouts` syntax

Provider versions before this change exposed `timeouts` as a **block** (`timeouts { delete = "20m" }`). The resource envelope now injects `timeouts` as an **attribute** with the same sub-fields. Update existing configuration:

```hcl
# Before (block)
timeouts {
  delete = "20m"
}

# After (attribute)
timeouts = {
  delete = "20m"
}
```

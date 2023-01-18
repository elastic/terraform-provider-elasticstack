# Cloud example

This example manages an elastic deployment in Elastic Cloud with an associated monitoring cluster. In addition, this example manages a handful of stack resources and demonstrates how these resources can be targeted to particular deployments. 

**In order to run this example, you first need to generate an API key in the Elastic Cloud console using the following instructions: https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html.**

```bash
terraform init
TF_VAR_ec_apikey=xxxxx terraform apply
```


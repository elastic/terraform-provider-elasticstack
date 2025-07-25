---
name: Bug report
about: Create a report to help us improve
title: "[Bug]"
labels: bug
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. TF configuration used '...'
2. TF operations to execute to get the error '...' [e.g `terraform plan`,`terraform apply`, `terraform destroy`]
3. See the error in the output '...'

**Expected behavior**
A clear and concise description of what you expected to happen.

**Debug output**
Run `terraform` command with `TF_LOG=trace` and provide extended information on TF operations. **Please ensure you redact any base64 encoded credentials from your output**. 
eg
```
[DEBUG] provider.terraform-provider-elasticstack_v0.11.0: Authorization: Basic xxxxx==
```

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Versions (please complete the following information):**
 - OS: [e.g. Linux]
 - Terraform Version [e.g. 1.0.0]
 - Provider version [e.g. v0.1.0]
 - Elasticsearch Version [e.g. 7.16.0]

**Additional context**
Add any other context about the problem here. Links to specific affected code files and paths here are also extremely useful (if known).

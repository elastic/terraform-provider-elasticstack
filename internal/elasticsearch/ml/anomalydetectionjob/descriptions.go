package anomalydetectionjob

import _ "embed"

//go:embed descriptions/resource.md
var resourceDescription string

//go:embed descriptions/job_id.md
var jobIDDescription string

//go:embed descriptions/bucket_span.md
var bucketSpanDescription string

//go:embed descriptions/by_field_name.md
var byFieldNameDescription string

//go:embed descriptions/over_field_name.md
var overFieldNameDescription string

//go:embed descriptions/per_partition_categorization_enabled.md
var perPartitionCategorizationEnabledDescription string

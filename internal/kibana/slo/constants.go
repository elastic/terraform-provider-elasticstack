package slo

import (
	"github.com/hashicorp/go-version"
)

var (
	SLOSupportsGroupByMinVersion                = version.Must(version.NewVersion("8.10.0"))
	SLOSupportsMultipleGroupByMinVersion        = version.Must(version.NewVersion("8.14.0"))
	SLOSupportsPreventInitialBackfillMinVersion = version.Must(version.NewVersion("8.15.0"))
	SLOSupportsDataViewIDMinVersion             = version.Must(version.NewVersion("8.15.0"))
)

// indicatorAddressToType maps Terraform block names to Kibana API indicator type strings.
var indicatorAddressToType = map[string]string{
	"apm_latency_indicator":      "sli.apm.transactionDuration",
	"apm_availability_indicator": "sli.apm.transactionErrorRate",
	"kql_custom_indicator":       "sli.kql.custom",
	"metric_custom_indicator":    "sli.metric.custom",
	"histogram_custom_indicator": "sli.histogram.custom",
	"timeslice_metric_indicator": "sli.metric.timeslice",
}

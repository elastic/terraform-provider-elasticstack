package models

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

type Slo struct {
	SloID           string
	Name            string
	Description     string
	Indicator       kbapi.SLOsSloDefinitionResponse_Indicator
	TimeWindow      kbapi.SLOsTimeWindow
	BudgetingMethod kbapi.SLOsBudgetingMethod
	Objective       kbapi.SLOsObjective
	Settings        *kbapi.SLOsSettings
	SpaceID         string
	GroupBy         []string
	Tags            []string
}

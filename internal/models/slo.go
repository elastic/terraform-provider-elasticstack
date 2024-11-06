package models

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
)

type Slo struct {
	SloID           string
	Name            string
	Description     string
	Indicator       slo.SloResponseIndicator
	TimeWindow      slo.TimeWindow
	BudgetingMethod slo.BudgetingMethod
	Objective       slo.Objective
	Settings        *slo.Settings
	SpaceID         string
	GroupBy         []string
	Tags            []string
}

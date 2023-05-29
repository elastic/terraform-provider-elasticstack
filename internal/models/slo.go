package models

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
)

type Slo struct {
	ID              string
	Name            string
	Description     string
	Indicator       slo.SloResponseIndicator
	TimeWindow      slo.SloResponseTimeWindow
	BudgetingMethod string //should I make this a slo.BudgetingMethod?
	Objective       slo.Objective
	Settings        *slo.Settings
	SpaceID         string
}

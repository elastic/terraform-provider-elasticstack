package contracttest

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// ParseDashboardPanel unmarshals JSON (one panel object from Kibana) into kbapi.DashboardPanelItem.
func ParseDashboardPanel(fullAPIResponse string) (kbapi.DashboardPanelItem, error) {
	var item kbapi.DashboardPanelItem
	if err := json.Unmarshal([]byte(fullAPIResponse), &item); err != nil {
		return kbapi.DashboardPanelItem{}, err
	}
	return item, nil
}

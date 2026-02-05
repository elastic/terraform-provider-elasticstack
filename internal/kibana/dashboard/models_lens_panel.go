package dashboard

import "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"

type lensPanelConfigConverter struct {
	visualizationType string
}

func (c lensPanelConfigConverter) handlesAPIPanelConfig(panelType string, cfg kbapi.DashboardPanelItem_Config) bool {
	if panelType != "lens" {
		return false
	}

	cfgMap, err := cfg.AsDashboardPanelItemConfig2()
	if err != nil {
		return false
	}

	return c.hasExpectedVisualizationType(cfgMap)
}

func (c lensPanelConfigConverter) hasExpectedVisualizationType(cfgMap map[string]interface{}) bool {
	attrs, ok := cfgMap["attributes"]
	if !ok {
		return false
	}

	attrsMap, ok := attrs.(map[string]interface{})
	if !ok {
		return false
	}

	vizType, ok := attrsMap["type"]
	if !ok {
		return false
	}

	return vizType == c.visualizationType
}

package dashboard

import "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"

type lensPanelConfigConverter struct {
	visualizationType string
	hasTFPanelConfig  func(pm panelModel) bool
}

func (c lensPanelConfigConverter) handlesAPIPanelConfig(pm *panelModel, panelType string, cfg kbapi.DashboardPanelItem_Config) bool {
	if c.hasTFPanelConfig != nil && pm != nil && !c.hasTFPanelConfig(*pm) {
		return false
	}

	if panelType != "lens" {
		return false
	}

	cfgMap, err := cfg.AsDashboardPanelItemConfig2()
	if err != nil {
		return false
	}

	return c.hasExpectedVisualizationType(cfgMap)
}

func (c lensPanelConfigConverter) hasExpectedVisualizationType(cfgMap map[string]any) bool {
	attrs, ok := cfgMap["attributes"]
	if !ok {
		return false
	}

	attrsMap, ok := attrs.(map[string]any)
	if !ok {
		return false
	}

	vizType, ok := attrsMap["type"]
	if !ok {
		return false
	}

	return vizType == c.visualizationType
}

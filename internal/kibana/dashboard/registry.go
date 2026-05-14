package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

// panelHandlers is populated as individual panel implementations land.
var panelHandlers = []iface.Handler{}

var panelTypeToHandler map[string]iface.Handler
var derivedPanelConfigNames []string

func init() {
	panelTypeToHandler = make(map[string]iface.Handler, len(panelHandlers))
	derivedPanelConfigNames = append(derivedPanelConfigNames, "config_json")
	for _, h := range panelHandlers {
		panelTypeToHandler[h.PanelType()] = h
		block := h.PanelType() + "_config"
		panelkit.MustPanelConfigBlockTagged(block)
		derivedPanelConfigNames = append(derivedPanelConfigNames, block)
	}
}

func LookupHandler(panelType string) iface.Handler { return panelTypeToHandler[panelType] }

func AllHandlers() []iface.Handler { return panelHandlers }

func ConfigNames() []string { return derivedPanelConfigNames }

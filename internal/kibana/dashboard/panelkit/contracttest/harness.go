package contracttest

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

// Config holds a raw dashboard panel fixture (one JSON object matching kbapi DashboardPanelItem).
type Config struct {
	FullAPIResponse string
	SkipFields      []string
}

func Run(t *testing.T, handler iface.Handler, cfg Config) {
	t.Helper()
	ctx := context.Background()
	for _, msg := range runChecks(ctx, handler, cfg) {
		t.Error(msg)
	}
}

func runChecks(ctx context.Context, handler iface.Handler, cfg Config) []string {
	block := handler.PanelType() + "_config"

	if cfg.FullAPIResponse == "" {
		return []string{"[Harness] FullAPIResponse must be non-empty JSON"}
	}
	if _, err := ParseDashboardPanel(cfg.FullAPIResponse); err != nil {
		return []string{"[Harness] parse FullAPIResponse: " + err.Error()}
	}

	var issues []string

	appendOuterSchemaIssues(handler, &issues)
	appendRequiredJSONPresenceIssues(handler, cfg.FullAPIResponse, &issues)
	if panelkit.HasPanelConfigBlock(block) {
		appendValidateRequiredZeroIssues(handler, cfg.FullAPIResponse, &issues)
	}

	appendRoundtripIssues(ctx, handler, cfg.FullAPIResponse, cfg.SkipFields, &issues)

	appendReflectIssues(ctx, handler, cfg.FullAPIResponse, &issues)

	appendNullPreserveIssues(ctx, handler, cfg.FullAPIResponse, cfg.SkipFields, &issues)

	return issues
}

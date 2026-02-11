package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func containsDataSource(list []func() datasource.DataSource, target func() datasource.DataSource) bool {
	targetPtr := reflect.ValueOf(target).Pointer()
	for _, entry := range list {
		if reflect.ValueOf(entry).Pointer() == targetPtr {
			return true
		}
	}
	return false
}

func TestFleetOutputDataSourceIsDefault(t *testing.T) {
	ctx := context.Background()

	t.Setenv(IncludeExperimentalEnvVar, "")
	provider := &Provider{version: "dev"}

	if !containsDataSource(provider.DataSources(ctx), output.NewDataSource) {
		t.Fatalf("expected fleet output data source to be enabled by default")
	}
}

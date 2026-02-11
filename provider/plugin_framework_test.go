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

func TestExperimentalDataSourceGate(t *testing.T) {
	ctx := context.Background()

	t.Run("default", func(t *testing.T) {
		t.Setenv(IncludeExperimentalEnvVar, "")
		provider := &Provider{version: "dev"}

		if containsDataSource(provider.DataSources(ctx), output.NewDataSource) {
			t.Fatalf("expected experimental data sources to be gated by default")
		}
	})

	t.Run("env-enabled", func(t *testing.T) {
		t.Setenv(IncludeExperimentalEnvVar, "true")
		provider := &Provider{version: "dev"}

		if !containsDataSource(provider.DataSources(ctx), output.NewDataSource) {
			t.Fatalf("expected experimental data sources to be enabled when %s=true", IncludeExperimentalEnvVar)
		}
	})

	t.Run("acctest-version", func(t *testing.T) {
		t.Setenv(IncludeExperimentalEnvVar, "")
		provider := &Provider{version: AccTestVersion}

		if !containsDataSource(provider.DataSources(ctx), output.NewDataSource) {
			t.Fatalf("expected experimental data sources to be enabled for acctest provider")
		}
	})
}

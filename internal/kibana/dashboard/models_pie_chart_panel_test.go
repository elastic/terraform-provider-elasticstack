package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pieChartConfigModel_fromAPI_toAPI_PieNoESQL(t *testing.T) {
	// Setup test data
	title := "My Pie Chart"
	desc := "A delicious pie chart"
	donutHole := kbapi.PieNoESQLDonutHoleSmall
	labelPos := kbapi.PieNoESQLLabelPositionInside

	// Create a dummy dataset
	dataset := kbapi.PieNoESQL_Dataset{}
	// Since dataset is a union or complex type in API usually, check if I need to marshal/unmarshal
	// Actually PieNoESQL_Dataset is likely a struct in generated code?
	// Let's assume it's simple enough for test.
	// To be safe, let's just use defaults or verify json marshalling works.

	visible := kbapi.PieLegendVisibleShow
	legend := kbapi.PieLegend{
		Visible: &visible,
	}

	query := kbapi.FilterSimpleSchema{
		Query:    "response:200",
		Language: utils.Pointer(kbapi.FilterSimpleSchemaLanguageKuery),
	}

	apiChart := kbapi.PieNoESQL{
		Title:         &title,
		Description:   &desc,
		DonutHole:     &donutHole,
		LabelPosition: &labelPos,
		Legend:        legend,
		Dataset:       dataset,
		Query:         query,
		Metrics:       []kbapi.PieNoESQL_Metrics_Item{}, // Empty for simplicity
		GroupBy:       utils.Pointer([]kbapi.PieNoESQL_GroupBy_Item{}),
	}

	// Wrap in PieChartSchema
	var apiSchema kbapi.PieChartSchema
	err := apiSchema.FromPieNoESQL(apiChart)
	require.NoError(t, err)

	// Test fromAPI
	ctx := context.Background()
	model := &pieChartConfigModel{}
	diags := model.fromAPI(ctx, apiSchema)
	require.False(t, diags.HasError(), "fromAPI should not have errors")

	// Verify fields
	assert.Equal(t, title, model.Title.ValueString())
	assert.Equal(t, desc, model.Description.ValueString())
	assert.Equal(t, string(donutHole), model.DonutHole.ValueString())
	assert.Equal(t, string(labelPos), model.LabelPosition.ValueString())
	assert.Equal(t, "response:200", model.Query.Query.ValueString())

	// Test toAPI
	resultSchema, diags := model.toAPI()
	require.False(t, diags.HasError(), "toAPI should not have errors")

	// Verify we can convert back to PieNoESQL
	resultNoESQL, err := resultSchema.AsPieNoESQL()
	require.NoError(t, err)

	assert.Equal(t, title, *resultNoESQL.Title)
	assert.Equal(t, desc, *resultNoESQL.Description)
}

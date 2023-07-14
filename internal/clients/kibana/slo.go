package kibana

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/slo"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func GetSlo(ctx context.Context, apiClient *clients.ApiClient, id, spaceID string) (*models.Slo, diag.Diagnostics) {
	client, err := apiClient.GetSloClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	req := client.GetSloOp(ctxWithAuth, "default", id).KbnXsrf("true") //fuck kibana spaces
	sloRes, res, err := req.Execute()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return sloResponseToModel("default", sloRes), utils.CheckHttpError(res, "Unable to get slo with ID "+string(id)) //fuck kibana spaces
}

func DeleteSlo(ctx context.Context, apiClient *clients.ApiClient, sloId string, spaceId string) diag.Diagnostics {
	client, err := apiClient.GetSloClient()
	if err != nil {
		return diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	req := client.DeleteSloOp(ctxWithAuth, sloId, spaceId).KbnXsrf("true")
	res, err := req.Execute()
	if err != nil && res == nil {
		return diag.FromErr(err)
	}

	defer res.Body.Close()
	return utils.CheckHttpError(res, "Unabled to delete slo with ID "+string(sloId))
}

func UpdateSlo(ctx context.Context, apiClient *clients.ApiClient, s models.Slo) (*models.Slo, diag.Diagnostics) {
	client, err := apiClient.GetSloClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	indicator, err := responseIndicatorToCreateSloRequestIndicator(s.Indicator)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	reqModel := slo.UpdateSloRequest{
		Name:            &s.Name,
		Description:     &s.Description,
		Indicator:       &indicator,
		TimeWindow:      &s.TimeWindow,
		BudgetingMethod: (*slo.BudgetingMethod)(&s.BudgetingMethod),
		Objective:       &s.Objective,
		Settings:        s.Settings,
	}

	req := client.UpdateSloOp(ctxWithAuth, s.SpaceID, s.ID).KbnXsrf("true").UpdateSloRequest(reqModel)
	slo, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if diags := utils.CheckHttpError(res, "Unable to update slo with id "+s.ID); diags.HasError() {
		return nil, diags
	}

	return sloResponseToModel(s.SpaceID, slo), diag.Diagnostics{}
}

func CreateSlo(ctx context.Context, apiClient *clients.ApiClient, s models.Slo) (*models.Slo, diag.Diagnostics) {
	client, err := apiClient.GetSloClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	indicator, err := responseIndicatorToCreateSloRequestIndicator(s.Indicator)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	reqModel := slo.CreateSloRequest{
		Name:            s.Name,
		Description:     s.Description,
		Indicator:       indicator,
		TimeWindow:      s.TimeWindow,
		BudgetingMethod: slo.BudgetingMethod(s.BudgetingMethod),
		Objective:       s.Objective,
		Settings:        s.Settings,
	}
	req := client.CreateSloOp(ctxWithAuth, s.SpaceID).KbnXsrf("true").CreateSloRequest(reqModel)
	sloRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()

	if diags := utils.CheckHttpError(res, "Unable to create slo"); diags.HasError() {
		return nil, diags
	}

	s.ID = sloRes.Id

	return &s, diag.Diagnostics{}
}

func responseIndicatorToCreateSloRequestIndicator(s slo.SloResponseIndicator) (slo.CreateSloRequestIndicator, error) {
	var ret slo.CreateSloRequestIndicator

	ind := s.GetActualInstance()
	switch ind.(type) {
	case *slo.IndicatorPropertiesApmAvailability:
		i, _ := ind.(*slo.IndicatorPropertiesApmAvailability)
		ret.IndicatorPropertiesApmAvailability = i

	case *slo.IndicatorPropertiesApmLatency:
		i, _ := ind.(*slo.IndicatorPropertiesApmLatency)
		ret.IndicatorPropertiesApmLatency = i

	case *slo.IndicatorPropertiesCustomKql:
		i, _ := ind.(*slo.IndicatorPropertiesCustomKql)
		ret.IndicatorPropertiesCustomKql = i

	case *slo.IndicatorPropertiesCustomMetric:
		i, _ := ind.(*slo.IndicatorPropertiesCustomMetric)
		ret.IndicatorPropertiesCustomMetric = i

	case *slo.IndicatorPropertiesHistogram:
		i, _ := ind.(*slo.IndicatorPropertiesHistogram)
		ret.IndicatorPropertiesHistogram = i

	default:
		return ret, fmt.Errorf("unknown indicator type: %T", ind)
	}

	return ret, nil
}

func sloResponseToModel(spaceID string, res *slo.SloResponse) *models.Slo {
	if res == nil {
		return nil
	}

	return &models.Slo{
		ID:              *res.Id,
		SpaceID:         spaceID,
		Name:            *res.Name,
		Description:     *res.Description,
		BudgetingMethod: string(*res.BudgetingMethod),
		Indicator:       *res.Indicator,
		TimeWindow:      *res.TimeWindow,
		Objective:       *res.Objective,
		Settings:        res.Settings,
	}
}

package kibana

import (
	"context"
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
	req := client.GetSlo(ctxWithAuth, id, spaceID)
	sloRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	return sloResponseToModel(spaceID, sloRes), utils.CheckHttpError(res, "Unable to get slo with ID "+string(id))
}

func DeleteSlo(ctx context.Context, apiClient *clients.ApiClient, sloId string, spaceId string) diag.Diagnostics {
	client, err := apiClient.GetSloClient()
	if err != nil {
		return diag.FromErr(err)
	}

	ctxWithAuth := apiClient.SetGeneratedClientAuthContext(ctx)
	req := client.DeleteSlo(ctxWithAuth, sloId, spaceId).KbnXsrf("true")
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
	reqModel := slo.UpdateSloRequest{
		Name:            &s.Name,
		Description:     &s.Description,
		Indicator:       (*slo.CreateSloRequestIndicator)(&s.Indicator),
		TimeWindow:      &s.TimeWindow,
		BudgetingMethod: (*slo.BudgetingMethod)(&s.BudgetingMethod),
		Objective:       &s.Objective,
		Settings:        s.Settings,
	}

	req := client.UpdateSlo(ctxWithAuth, s.ID, s.SpaceID).KbnXsrf("true").UpdateSloRequest(reqModel)
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

	reqModel := slo.CreateSloRequest{
		Name:            s.Name,
		Description:     s.Description,
		Indicator:       slo.CreateSloRequestIndicator(s.Indicator),
		TimeWindow:      s.TimeWindow,
		BudgetingMethod: slo.BudgetingMethod(s.BudgetingMethod),
		Objective:       s.Objective,
	}
	req := client.CreateSlo(ctxWithAuth, s.SpaceID).KbnXsrf("true").CreateSloRequest(reqModel)
	sloRes, res, err := req.Execute()
	if err != nil && res == nil {
		return nil, diag.FromErr(err)
	}

	s.ID = sloRes.Id
	defer res.Body.Close()

	return &s, diag.Diagnostics{}
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

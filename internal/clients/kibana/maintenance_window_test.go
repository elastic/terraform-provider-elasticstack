package kibana

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/alerting"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func Test_maintenanceWindowResponseToModel(t *testing.T) {
	tests := []struct {
		name                      string
		spaceId                   string
		maintenanceWindowResponse *alerting.MaintenanceWindowResponseProperties
		expectedModel             *models.MaintenanceWindow
	}{
		{
			name:                      "nil response should return a nil model",
			spaceId:                   "space-id",
			maintenanceWindowResponse: nil,
			expectedModel:             nil,
		},
		{
			name:    "nil optional fields should not blow up the transform",
			spaceId: "space-id",
			maintenanceWindowResponse: &alerting.MaintenanceWindowResponseProperties{
				Id:        "some-long-id",
				CreatedBy: "me",
				CreatedAt: "today",
				UpdatedBy: "me",
				UpdatedAt: "today",
				Enabled:   true,
				Status:    "running",
				Title:     "maintenance-window-id",
				Schedule: alerting.MaintenanceWindowResponsePropertiesSchedule{
					Custom: alerting.MaintenanceWindowResponsePropertiesScheduleCustom{
						Duration:  "12d",
						Start:     "1999-02-02T05:00:00.200Z",
						Recurring: nil,
						Timezone:  nil,
					},
				},
				Scope: nil,
			},
			expectedModel: &models.MaintenanceWindow{
				MaintenanceWindowId: "some-long-id",
				SpaceId:             "space-id",
				Title:               "maintenance-window-id",
				Enabled:             true,
				CustomSchedule: models.MaintenanceWindowSchedule{
					Duration: "12d",
					Start:    "1999-02-02T05:00:00.200Z",
				},
			},
		},
		{
			name:    "a full response should be successfully transformed",
			spaceId: "space-id",
			maintenanceWindowResponse: &alerting.MaintenanceWindowResponseProperties{
				Id:        "maintenance-window-id",
				Title:     "maintenance-window-title",
				CreatedBy: "me",
				CreatedAt: "today",
				UpdatedBy: "me",
				UpdatedAt: "today",
				Enabled:   true,
				Status:    "running",
				Schedule: alerting.MaintenanceWindowResponsePropertiesSchedule{
					Custom: alerting.MaintenanceWindowResponsePropertiesScheduleCustom{
						Duration: "12d",
						Start:    "1999-02-02T05:00:00.200Z",
						Timezone: utils.Pointer("Asia/Taipei"),
						Recurring: &alerting.MaintenanceWindowResponsePropertiesScheduleCustomRecurring{
							End:        utils.Pointer("2029-05-17T05:05:00.000Z"),
							Every:      utils.Pointer("20d"),
							OnMonth:    []float32{2},
							OnMonthDay: []float32{1},
							OnWeekDay:  []string{"WE", "TU"},
						},
					},
				},
				Scope: &alerting.MaintenanceWindowResponsePropertiesScope{
					Alerting: alerting.MaintenanceWindowResponsePropertiesScopeAlerting{
						Query: alerting.MaintenanceWindowResponsePropertiesScopeAlertingQuery{
							Kql: "_id: 'foobar'",
						},
					},
				},
			},
			expectedModel: &models.MaintenanceWindow{
				MaintenanceWindowId: "maintenance-window-id",
				Title:               "maintenance-window-title",
				SpaceId:             "space-id",
				Enabled:             true,
				CustomSchedule: models.MaintenanceWindowSchedule{
					Duration: "12d",
					Start:    "1999-02-02T05:00:00.200Z",
					Timezone: utils.Pointer("Asia/Taipei"),
					Recurring: &models.MaintenanceWindowScheduleRecurring{
						End:        utils.Pointer("2029-05-17T05:05:00.000Z"),
						Every:      utils.Pointer("20d"),
						OnMonth:    &[]float32{2},
						OnMonthDay: &[]float32{1},
						OnWeekDay:  &[]string{"WE", "TU"},
					},
				},
				Scope: &models.MaintenanceWindowScope{
					Alerting: &models.MaintenanceWindowAlertingScope{
						Kql: "_id: 'foobar'",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := maintenanceWindowResponseToModel(tt.spaceId, tt.maintenanceWindowResponse)

			require.Equal(t, tt.expectedModel, model)
		})
	}
}

func Test_CreateUpdateMaintenanceWindow(t *testing.T) {
	ctrl := gomock.NewController(t)

	getApiClient := func() (ApiClient, *alerting.MockAlertingAPI) {
		apiClient := NewMockApiClient(ctrl)
		apiClient.EXPECT().SetAlertingAuthContext(gomock.Any()).DoAndReturn(func(ctx context.Context) context.Context { return ctx })
		alertingClient := alerting.NewMockAlertingAPI(ctrl)
		apiClient.EXPECT().GetAlertingClient().DoAndReturn(func() (alerting.AlertingAPI, error) { return alertingClient, nil })
		return apiClient, alertingClient
	}

	tests := []struct {
		name              string
		testFunc          func(ctx context.Context, apiClient ApiClient, maintenanceWindow models.MaintenanceWindow) (*models.MaintenanceWindow, diag.Diagnostics)
		client            ApiClient
		maintenanceWindow models.MaintenanceWindow
		expectedRes       *models.MaintenanceWindow
		expectedErr       string
	}{
		{
			name:     "CreateMaintenanceWindow should not crash when backend returns 4xx",
			testFunc: CreateMaintenanceWindow,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().CreateMaintenanceWindow(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, spaceId string) alerting.ApiCreateMaintenanceWindowRequest {
					return alerting.ApiCreateMaintenanceWindowRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().CreateMaintenanceWindowExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiCreateMaintenanceWindowRequest) (*alerting.MaintenanceWindowResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 401,
						Body:       io.NopCloser(strings.NewReader("some error")),
					}, nil
				})
				return apiClient
			}(),
			maintenanceWindow: models.MaintenanceWindow{},
			expectedRes:       nil,
			expectedErr:       "some error",
		},
		{
			name:     "UpdateMaintenanceWindow should not crash when backend returns 4xx",
			testFunc: UpdateMaintenanceWindow,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().UpdateMaintenanceWindow(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, maintenanceWindowId string, spaceId string) alerting.ApiUpdateMaintenanceWindowRequest {
					return alerting.ApiUpdateMaintenanceWindowRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().UpdateMaintenanceWindowExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiUpdateMaintenanceWindowRequest) (*alerting.MaintenanceWindowResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 401,
						Body:       io.NopCloser(strings.NewReader("some error")),
					}, nil
				})
				return apiClient
			}(),
			maintenanceWindow: models.MaintenanceWindow{},
			expectedRes:       nil,
			expectedErr:       "some error",
		},
		{
			name:     "CreateMaintenanceWindow should not crash when backend returns an empty response and HTTP 200",
			testFunc: CreateMaintenanceWindow,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().CreateMaintenanceWindow(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, spaceId string) alerting.ApiCreateMaintenanceWindowRequest {
					return alerting.ApiCreateMaintenanceWindowRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().CreateMaintenanceWindowExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiCreateMaintenanceWindowRequest) (*alerting.MaintenanceWindowResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("everything seems fine")),
					}, nil
				})
				return apiClient
			}(),
			maintenanceWindow: models.MaintenanceWindow{},
			expectedRes:       nil,
			expectedErr:       "empty response",
		},
		{
			name:     "UpdateMaintenanceWindow should not crash when backend returns an empty response and HTTP 200",
			testFunc: UpdateMaintenanceWindow,
			client: func() ApiClient {
				apiClient, alertingClient := getApiClient()
				alertingClient.EXPECT().UpdateMaintenanceWindow(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, maintenanceWindowId string, spaceId string) alerting.ApiUpdateMaintenanceWindowRequest {
					return alerting.ApiUpdateMaintenanceWindowRequest{ApiService: alertingClient}
				})
				alertingClient.EXPECT().UpdateMaintenanceWindowExecute(gomock.Any()).DoAndReturn(func(r alerting.ApiUpdateMaintenanceWindowRequest) (*alerting.MaintenanceWindowResponseProperties, *http.Response, error) {
					return nil, &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("everything seems fine")),
					}, nil
				})
				return apiClient
			}(),
			maintenanceWindow: models.MaintenanceWindow{},
			expectedRes:       nil,
			expectedErr:       "empty response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maintenanceWindow, diags := tt.testFunc(context.Background(), tt.client, tt.maintenanceWindow)

			if tt.expectedRes == nil {
				require.Nil(t, maintenanceWindow)
			} else {
				require.Equal(t, tt.expectedRes, maintenanceWindow)
			}

			if tt.expectedErr != "" {
				require.NotEmpty(t, diags)
				if !strings.Contains(diags[0].Detail, tt.expectedErr) {
					require.Fail(t, fmt.Sprintf("Diags ['%s'] should contain message ['%s']", diags[0].Detail, tt.expectedErr))
				}
			}
		})
	}
}

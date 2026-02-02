package dashboard

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_markdownPanelConfigConverter_handlesAPIPanelConfig(t *testing.T) {
	tests := []struct {
		name      string
		panelType string
		want      bool
	}{
		{
			name:      "handles DASHBOARD_MARKDOWN type",
			panelType: "DASHBOARD_MARKDOWN",
			want:      true,
		},
		{
			name:      "does not handle visualization type",
			panelType: "visualization",
			want:      false,
		},
		{
			name:      "does not handle lens type",
			panelType: "lens",
			want:      false,
		},
		{
			name:      "does not handle search type",
			panelType: "search",
			want:      false,
		},
		{
			name:      "does not handle empty string",
			panelType: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := markdownPanelConfigConverter{}
			got := c.handlesAPIPanelConfig(tt.panelType, kbapi.DashboardPanelItem_Config{})
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_markdownPanelConfigConverter_populateFromAPIPanel(t *testing.T) {
	tests := []struct {
		name        string
		config      kbapi.DashboardPanelItem_Config
		expected    *markdownConfigModel
		expectError bool
	}{
		{
			name: "all fields populated",
			config: func() kbapi.DashboardPanelItem_Config {
				content := "# Markdown Content"
				description := "A test description"
				hidePanelTitles := true
				title := "My Panel Title"

				config0 := kbapi.DashboardPanelItemConfig0{
					Content:         content,
					Description:     &description,
					HidePanelTitles: &hidePanelTitles,
					Title:           &title,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue("# Markdown Content"),
				Description:     types.StringValue("A test description"),
				HidePanelTitles: types.BoolValue(true),
				Title:           types.StringValue("My Panel Title"),
			},
			expectError: false,
		},
		{
			name: "only required field (content)",
			config: func() kbapi.DashboardPanelItem_Config {
				config0 := kbapi.DashboardPanelItemConfig0{
					Content:         "Simple content",
					Description:     nil,
					HidePanelTitles: nil,
					Title:           nil,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue("Simple content"),
				Description:     types.StringNull(),
				HidePanelTitles: types.BoolNull(),
				Title:           types.StringNull(),
			},
			expectError: false,
		},
		{
			name: "empty content string",
			config: func() kbapi.DashboardPanelItem_Config {
				config0 := kbapi.DashboardPanelItemConfig0{
					Content: "",
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue(""),
				Description:     types.StringNull(),
				HidePanelTitles: types.BoolNull(),
				Title:           types.StringNull(),
			},
			expectError: false,
		},
		{
			name: "hidePanelTitles is false",
			config: func() kbapi.DashboardPanelItem_Config {
				hidePanelTitles := false
				config0 := kbapi.DashboardPanelItemConfig0{
					Content:         "Content",
					HidePanelTitles: &hidePanelTitles,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue("Content"),
				Description:     types.StringNull(),
				HidePanelTitles: types.BoolValue(false),
				Title:           types.StringNull(),
			},
			expectError: false,
		},
		{
			name: "empty description and title strings",
			config: func() kbapi.DashboardPanelItem_Config {
				description := ""
				title := ""
				config0 := kbapi.DashboardPanelItemConfig0{
					Content:     "Content",
					Description: &description,
					Title:       &title,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue("Content"),
				Description:     types.StringValue(""),
				HidePanelTitles: types.BoolNull(),
				Title:           types.StringValue(""),
			},
			expectError: false,
		},
		{
			name: "multiline markdown content",
			config: func() kbapi.DashboardPanelItem_Config {
				content := `# Header
## Subheader

Some **bold** text and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`
				config0 := kbapi.DashboardPanelItemConfig0{
					Content: content,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content: types.StringValue(`# Header
## Subheader

Some **bold** text and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`),
				Description:     types.StringNull(),
				HidePanelTitles: types.BoolNull(),
				Title:           types.StringNull(),
			},
			expectError: false,
		},
		{
			name: "special characters in content",
			config: func() kbapi.DashboardPanelItem_Config {
				content := `Content with special chars: <>&"'`
				config0 := kbapi.DashboardPanelItemConfig0{
					Content: content,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue(`Content with special chars: <>&"'`),
				Description:     types.StringNull(),
				HidePanelTitles: types.BoolNull(),
				Title:           types.StringNull(),
			},
			expectError: false,
		},
		{
			name: "config with additional unknown fields still works",
			config: func() kbapi.DashboardPanelItem_Config {
				// Even with extra fields, as long as required fields exist, it should work
				content := "Content with extra fields"
				description := "Description"
				config0 := kbapi.DashboardPanelItemConfig0{
					Content:     content,
					Description: &description,
				}

				config := kbapi.DashboardPanelItem_Config{}
				err := config.FromDashboardPanelItemConfig0(config0)
				if err != nil {
					panic(err)
				}
				return config
			}(),
			expected: &markdownConfigModel{
				Content:         types.StringValue("Content with extra fields"),
				Description:     types.StringValue("Description"),
				HidePanelTitles: types.BoolNull(),
				Title:           types.StringNull(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := markdownPanelConfigConverter{}
			pm := &panelModel{}
			ctx := context.Background()

			diags := c.populateFromAPIPanel(ctx, pm, tt.config)

			if tt.expectError {
				require.True(t, diags.HasError(), "expected error but got none")
				return
			}

			require.False(t, diags.HasError(), "unexpected error: %v", diags)
			require.NotNil(t, pm.MarkdownConfig, "MarkdownConfig should not be nil")
			assert.Equal(t, tt.expected.Content, pm.MarkdownConfig.Content, "Content mismatch")
			assert.Equal(t, tt.expected.Description, pm.MarkdownConfig.Description, "Description mismatch")
			assert.Equal(t, tt.expected.HidePanelTitles, pm.MarkdownConfig.HidePanelTitles, "HidePanelTitles mismatch")
			assert.Equal(t, tt.expected.Title, pm.MarkdownConfig.Title, "Title mismatch")
		})
	}
}

func Test_markdownPanelConfigConverter_populateFromAPIPanel_contextsAreHandled(t *testing.T) {
	// Test that the context is properly handled (even though currently not used in the implementation)
	c := markdownPanelConfigConverter{}
	pm := &panelModel{}

	config0 := kbapi.DashboardPanelItemConfig0{
		Content: "Test content",
	}

	config := kbapi.DashboardPanelItem_Config{}
	err := config.FromDashboardPanelItemConfig0(config0)
	require.NoError(t, err)

	// Test with background context
	ctx := context.Background()
	diags := c.populateFromAPIPanel(ctx, pm, config)
	require.False(t, diags.HasError())
	require.NotNil(t, pm.MarkdownConfig)

	// Test with canceled context (should still work as context is not used currently)
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	pm2 := &panelModel{}
	diags2 := c.populateFromAPIPanel(canceledCtx, pm2, config)
	require.False(t, diags2.HasError())
	require.NotNil(t, pm2.MarkdownConfig)
}

func Test_markdownPanelConfigConverter_handlesTFPanelConfig(t *testing.T) {
	tests := []struct {
		name string
		pm   panelModel
		want bool
	}{
		{
			name: "handles panel with markdown config",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content: types.StringValue("# Hello"),
				},
			},
			want: true,
		},
		{
			name: "does not handle panel without markdown config",
			pm: panelModel{
				MarkdownConfig: nil,
			},
			want: false,
		},
		{
			name: "does not handle panel with other config types",
			pm: panelModel{
				XYChartConfig: &xyChartConfigModel{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := markdownPanelConfigConverter{}
			got := c.handlesTFPanelConfig(tt.pm)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_markdownPanelConfigConverter_mapPanelToAPI(t *testing.T) {
	tests := []struct {
		name       string
		pm         panelModel
		wantConfig kbapi.DashboardPanelItemConfig0
		wantDiags  bool
	}{
		{
			name: "successfully maps panel with all fields to API config",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content:         types.StringValue("# Test Content"),
					Description:     types.StringValue("Test Description"),
					HidePanelTitles: types.BoolValue(true),
					Title:           types.StringValue("Test Title"),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content:         "# Test Content",
				Description:     strPtr("Test Description"),
				HidePanelTitles: boolPtr(true),
				Title:           strPtr("Test Title"),
			},
			wantDiags: false,
		},
		{
			name: "successfully maps panel with minimal fields to API config",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content:         types.StringValue("# Minimal"),
					Description:     types.StringNull(),
					HidePanelTitles: types.BoolNull(),
					Title:           types.StringNull(),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content:         "# Minimal",
				Description:     nil,
				HidePanelTitles: nil,
				Title:           nil,
			},
			wantDiags: false,
		},
		{
			name: "successfully maps panel with unknown values",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content:         types.StringValue("# Content"),
					Description:     types.StringUnknown(),
					HidePanelTitles: types.BoolUnknown(),
					Title:           types.StringUnknown(),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content:         "# Content",
				Description:     nil,
				HidePanelTitles: nil,
				Title:           nil,
			},
			wantDiags: false,
		},
		{
			name: "handles false value for hidePanelTitles",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content:         types.StringValue("# Content"),
					HidePanelTitles: types.BoolValue(false),
					Description:     types.StringNull(),
					Title:           types.StringNull(),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content:         "# Content",
				HidePanelTitles: boolPtr(false),
				Description:     nil,
				Title:           nil,
			},
			wantDiags: false,
		},
		{
			name: "handles empty string values",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content:     types.StringValue(""),
					Description: types.StringValue(""),
					Title:       types.StringValue(""),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content:     "",
				Description: strPtr(""),
				Title:       strPtr(""),
			},
			wantDiags: false,
		},
		{
			name: "handles multiline markdown content",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content: types.StringValue(`# Header
## Subheader

Some **bold** text and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`),
					Description:     types.StringNull(),
					HidePanelTitles: types.BoolNull(),
					Title:           types.StringNull(),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content: `# Header
## Subheader

Some **bold** text and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`,
				Description:     nil,
				HidePanelTitles: nil,
				Title:           nil,
			},
			wantDiags: false,
		},
		{
			name: "handles special characters",
			pm: panelModel{
				MarkdownConfig: &markdownConfigModel{
					Content:     types.StringValue(`Content with special chars: <>&"'`),
					Description: types.StringValue(`Description with special chars: <>&"'`),
					Title:       types.StringValue(`Title with special chars: <>&"'`),
				},
			},
			wantConfig: kbapi.DashboardPanelItemConfig0{
				Content:     `Content with special chars: <>&"'`,
				Description: strPtr(`Description with special chars: <>&"'`),
				Title:       strPtr(`Title with special chars: <>&"'`),
			},
			wantDiags: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := markdownPanelConfigConverter{}
			apiConfig := kbapi.DashboardPanelItem_Config{}

			diags := c.mapPanelToAPI(tt.pm, &apiConfig)

			if tt.wantDiags {
				assert.True(t, diags.HasError(), "expected diagnostics to have errors")
			} else {
				assert.False(t, diags.HasError(), "expected no diagnostics errors")

				// Extract the config and verify
				config0, err := apiConfig.AsDashboardPanelItemConfig0()
				require.NoError(t, err, "failed to extract config")

				assert.Equal(t, tt.wantConfig.Content, config0.Content)

				if tt.wantConfig.Description == nil {
					assert.Nil(t, config0.Description)
				} else {
					require.NotNil(t, config0.Description)
					assert.Equal(t, *tt.wantConfig.Description, *config0.Description)
				}

				if tt.wantConfig.HidePanelTitles == nil {
					assert.Nil(t, config0.HidePanelTitles)
				} else {
					require.NotNil(t, config0.HidePanelTitles)
					assert.Equal(t, *tt.wantConfig.HidePanelTitles, *config0.HidePanelTitles)
				}

				if tt.wantConfig.Title == nil {
					assert.Nil(t, config0.Title)
				} else {
					require.NotNil(t, config0.Title)
					assert.Equal(t, *tt.wantConfig.Title, *config0.Title)
				}
			}
		})
	}
}

// Helper functions for pointer values
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

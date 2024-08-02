package synthetics

import (
	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	MetadataPrefix = "_kibana_synthetics_"
)

func MonitorIDSchema() schema.Attribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Generated identifier for the monitor",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func MonitorScheduleSchema() schema.Attribute {
	return schema.Int64Attribute{
		Optional:            true,
		Computed:            true,
		MarkdownDescription: "(Optional, number): The monitor’s schedule in minutes. Supported values are 1, 3, 5, 10, 15, 30, 60, 120 and 240.",
		Validators: []validator.Int64{
			int64validator.OneOf(1, 3, 5, 10, 15, 30, 60, 120, 240),
		},
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	}
}

func JsonObjectSchema() schema.Attribute {
	return schema.StringAttribute{
		Computed: true,
		Optional: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		MarkdownDescription: "Raw JSON object, use `jsonencode` function to represent JSON",
	}
}

func StatusConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "",
			},
		},
	}
}

func MonitorAlertConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "",
		Attributes: map[string]schema.Attribute{
			"status": StatusConfigSchema(),
			"tls":    StatusConfigSchema(),
		},
	}
}

func MonitorConfigSchema() schema.Attribute {
	//TODO: add MonitorIDSchema
	return schema.SingleNestedAttribute{
		Optional:    false,
		Description: "",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    false,
				Description: "",
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"schedule": MonitorScheduleSchema(),
			"locations": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "",
			},
			"private_locations": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "",
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "",
			},
			"alert": MonitorAlertConfigSchema(),
			"service_name": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "",
			},
			"namespace": schema.StringAttribute{
				Optional:    true,
				Description: "",
			},
			"params": JsonObjectSchema(),
			"retest_on_failure": schema.BoolAttribute{
				Optional:    true,
				Description: "",
			},
		},
	}
}

func MonitorScheduleConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    false,
		Description: "",
		Attributes: map[string]schema.Attribute{
			"number": schema.StringAttribute{
				Optional:    false,
				Description: "",
			},
			"unit": schema.StringAttribute{
				Optional:    false,
				Description: "",
			},
		},
	}
}

func GeoConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Geographic coordinates (WGS84) for the location",
		Attributes: map[string]schema.Attribute{
			"lat": schema.Float64Attribute{
				Optional:            false,
				MarkdownDescription: "",
			},
			"lon": schema.Float64Attribute{
				Optional:            false,
				MarkdownDescription: "",
			},
		},
	}
}

func MonitorLocationConfigSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    false,
		Description: "",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    false,
				Description: "",
			},
			"label": schema.StringAttribute{
				Optional:    false,
				Description: "",
			},
			"geo": GeoConfigSchema(),
			"is_service_managed": schema.BoolAttribute{
				Optional:    false,
				Description: "",
			},
		},
	}
}

type TFGeoConfigV0 struct {
	Lat types.Float64 `tfsdk:"lat"`
	Lon types.Float64 `tfsdk:"lon"`
}

func (m *TFGeoConfigV0) ToSyntheticGeoConfig() kbapi.SyntheticGeoConfig {
	return kbapi.SyntheticGeoConfig{
		Lat: m.Lat.ValueFloat64(),
		Lon: m.Lon.ValueFloat64(),
	}
}

func FromSyntheticGeoConfig(v kbapi.SyntheticGeoConfig) TFGeoConfigV0 {
	return TFGeoConfigV0{
		Lat: types.Float64Value(v.Lat),
		Lon: types.Float64Value(v.Lon),
	}
}

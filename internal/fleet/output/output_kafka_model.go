package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type outputKafkaModel struct {
	AuthType         types.String  `tfsdk:"auth_type"`
	BrokerTimeout    types.Float32 `tfsdk:"broker_timeout"`
	ClientId         types.String  `tfsdk:"client_id"`
	Compression      types.String  `tfsdk:"compression"`
	CompressionLevel types.Int64   `tfsdk:"compression_level"`
	ConnectionType   types.String  `tfsdk:"connection_type"`
	Topic            types.String  `tfsdk:"topic"`
	Partition        types.String  `tfsdk:"partition"`
	RequiredAcks     types.Int64   `tfsdk:"required_acks"`
	Timeout          types.Float32 `tfsdk:"timeout"`
	Version          types.String  `tfsdk:"version"`
	Username         types.String  `tfsdk:"username"`
	Password         types.String  `tfsdk:"password"`
	Key              types.String  `tfsdk:"key"`
	Headers          types.List    `tfsdk:"headers"`     //> outputHeadersModel
	Hash             types.Object  `tfsdk:"hash"`        //> outputHashModel
	Random           types.Object  `tfsdk:"random"`      //> outputRandomModel
	RoundRobin       types.Object  `tfsdk:"round_robin"` //> outputRoundRobinModel
	Sasl             types.Object  `tfsdk:"sasl"`        //> outputSaslModel
}

type outputHeadersModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type outputHashModel struct {
	Hash   types.String `tfsdk:"hash"`
	Random types.Bool   `tfsdk:"random"`
}

type outputRandomModel struct {
	GroupEvents types.Float64 `tfsdk:"group_events"`
}

type outputRoundRobinModel struct {
	GroupEvents types.Float64 `tfsdk:"group_events"`
}

type outputSaslModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
}

func (m outputKafkaModel) toAPIHash(ctx context.Context) (*struct {
	Hash   *string `json:"hash,omitempty"`
	Random *bool   `json:"random,omitempty"`
}, diag.Diagnostics) {
	if !utils.IsKnown(m.Hash) {
		return nil, nil
	}

	var hashModel outputHashModel
	diags := m.Hash.As(ctx, &hashModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &struct {
		Hash   *string `json:"hash,omitempty"`
		Random *bool   `json:"random,omitempty"`
	}{
		Hash:   hashModel.Hash.ValueStringPointer(),
		Random: hashModel.Random.ValueBoolPointer(),
	}, diags
}

func (m outputKafkaModel) toAPIHeaders(ctx context.Context) (*[]struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}, diag.Diagnostics) {
	if !utils.IsKnown(m.Headers) {
		return nil, nil
	}

	var diags diag.Diagnostics
	headerModels := utils.ListTypeAs[outputHeadersModel](ctx, m.Headers, path.Root("kafka").AtName("headers"), &diags)
	if len(headerModels) == 0 {
		return nil, diags
	}

	headers := make([]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}, len(headerModels))
	for i, h := range headerModels {
		headers[i] = struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{
			Key:   h.Key.ValueString(),
			Value: h.Value.ValueString(),
		}
	}
	return &headers, diags
}

func (m outputKafkaModel) toAPIRandom(ctx context.Context) (*struct {
	GroupEvents *float32 `json:"group_events,omitempty"`
}, diag.Diagnostics) {
	if !utils.IsKnown(m.Random) {
		return nil, nil
	}

	var randomModel outputRandomModel
	diags := m.Random.As(ctx, &randomModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &struct {
		GroupEvents *float32 `json:"group_events,omitempty"`
	}{
		GroupEvents: func() *float32 {
			if !randomModel.GroupEvents.IsNull() {
				val := float32(randomModel.GroupEvents.ValueFloat64())
				return &val
			}
			return nil
		}(),
	}, diags
}

func (m outputKafkaModel) toAPIRoundRobin(ctx context.Context) (*struct {
	GroupEvents *float32 `json:"group_events,omitempty"`
}, diag.Diagnostics) {
	if !utils.IsKnown(m.RoundRobin) {
		return nil, nil
	}

	var roundRobinModel outputRoundRobinModel
	diags := m.RoundRobin.As(ctx, &roundRobinModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	return &struct {
		GroupEvents *float32 `json:"group_events,omitempty"`
	}{
		GroupEvents: func() *float32 {
			if !roundRobinModel.GroupEvents.IsNull() {
				val := float32(roundRobinModel.GroupEvents.ValueFloat64())
				return &val
			}
			return nil
		}(),
	}, nil
}

func (m outputKafkaModel) toAPISasl(ctx context.Context) (*struct {
	Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
}, diag.Diagnostics) {
	if !utils.IsKnown(m.Sasl) {
		return nil, nil
	}
	var saslModel outputSaslModel
	diags := m.Sasl.As(ctx, &saslModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	if saslModel.Mechanism.IsNull() {
		return nil, diags
	}

	mechanism := kbapi.NewOutputKafkaSaslMechanism(saslModel.Mechanism.ValueString())
	return &struct {
		Mechanism *kbapi.NewOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
	}{
		Mechanism: &mechanism,
	}, diags
}

func (m outputKafkaModel) toUpdateAPISasl(ctx context.Context) (*struct {
	Mechanism *kbapi.UpdateOutputKafkaSaslMechanism `json:"mechanism,omitempty"`
}, diag.Diagnostics) {
	sasl, diags := m.toAPISasl(ctx)
	if diags.HasError() || sasl == nil {
		return nil, diags
	}

	mechanism := kbapi.UpdateOutputKafkaSaslMechanism(*sasl.Mechanism)
	return &struct {
		Mechanism *kbapi.UpdateOutputKafkaSaslMechanism "json:\"mechanism,omitempty\""
	}{
		Mechanism: &mechanism,
	}, diags
}

func (m outputKafkaModel) toAuthType() kbapi.NewOutputKafkaAuthType {
	if !utils.IsKnown(m.AuthType) {
		return kbapi.NewOutputKafkaAuthTypeNone
	}

	return kbapi.NewOutputKafkaAuthType(m.AuthType.ValueString())
}

func (m outputKafkaModel) toUpdateAuthType() *kbapi.UpdateOutputKafkaAuthType {
	if !utils.IsKnown(m.AuthType) {
		return nil
	}

	return utils.Pointer(kbapi.UpdateOutputKafkaAuthType(m.AuthType.ValueString()))
}

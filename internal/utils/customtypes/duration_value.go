package customtypes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*Duration)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*Duration)(nil)
	_ xattr.ValidateableAttribute                = (*Duration)(nil)
)

type Duration struct {
	basetypes.StringValue
}

// Type returns a DurationType.
func (v Duration) Type(_ context.Context) attr.Type {
	return DurationType{}
}

// Equal returns true if the given value is equivalent.
func (v Duration) Equal(o attr.Value) bool {
	other, ok := o.(Duration)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (t Duration) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if t.IsNull() || t.IsUnknown() {
		return
	}

	valueString := t.ValueString()
	if _, err := time.ParseDuration(valueString); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Duration string value",
			fmt.Sprintf(`A string value was provided that is not a valid Go duration\n\nGiven value "%s"\n`, valueString),
		)
	}
}

// StringSemanticEquals returns true if the given duration string value is semantically equal to the current duration string value.
// When compared, the durations are parsed into a time.Duration and the underlying nanosecond values compared.
func (v Duration) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(Duration)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	vParsed, diags := v.Parse()
	if diags.HasError() {
		return false, diags
	}

	newParsed, diags := newValue.Parse()
	if diags.HasError() {
		return false, diags
	}

	return vParsed == newParsed, diags
}

// Parse calls time.ParseDuration with the Duration StringValue. A null or unknown value will produce an error diagnostic.
func (v Duration) Parse() (time.Duration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("Duration Parse error", "duration string value is null"))
		return 0, diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("Duration Parse Error", "duration string value is unknown"))
		return 0, diags
	}

	duration, err := time.ParseDuration(v.ValueString())
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Duration Parse Error", err.Error()))
	}

	return duration, diags
}

// NewDurationNull creates a Duration with a null value. Determine whether the value is null via IsNull method.
func NewDurationNull() Duration {
	return Duration{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewDurationUnknown creates a Duration with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewDurationUnknown() Duration {
	return Duration{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewDurationValue creates a Duration with a known value. Access the value via ValueString method.
func NewDurationValue(value string) Duration {
	return Duration{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewDurationPointerValue creates a Duration with a null value if nil or a known value. Access the value via ValueStringPointer method.
func NewDurationPointerValue(value *string) Duration {
	return Duration{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}

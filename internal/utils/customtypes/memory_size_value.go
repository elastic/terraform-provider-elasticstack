package customtypes

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.StringValuable                   = (*MemorySize)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*MemorySize)(nil)
	_ xattr.ValidateableAttribute                = (*MemorySize)(nil)
)

// memoryPattern matches memory size strings with optional 'b' suffix
var memoryPattern = regexp.MustCompile(`^(\d+)([kmgtKMGT])?[bB]?$`)

type MemorySize struct {
	basetypes.StringValue
}

// Type returns a MemorySizeType.
func (v MemorySize) Type(_ context.Context) attr.Type {
	return MemorySizeType{}
}

// Equal returns true if the given value is equivalent.
func (v MemorySize) Equal(o attr.Value) bool {
	other, ok := o.(MemorySize)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (t MemorySize) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if t.IsNull() || t.IsUnknown() {
		return
	}

	valueString := t.ValueString()
	if !memoryPattern.MatchString(valueString) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid memory size string value",
			fmt.Sprintf("A string value was provided that is not a valid memory size format\n\nGiven value \"%s\"\nExpected format: number followed by optional unit (k/K, m/M, g/G, t/T) and optional 'b/B' suffix", valueString),
		)
	}
}

// StringSemanticEquals returns true if the given memory size string value is semantically equal to the current memory size string value.
// When compared, the memory sizes are parsed into bytes and the byte values compared.
func (v MemorySize) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(MemorySize)
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

	vParsed, diags := v.ConvertToMB()
	if diags.HasError() {
		return false, diags
	}

	newParsed, diags := newValue.ConvertToMB()
	if diags.HasError() {
		return false, diags
	}

	return vParsed == newParsed, diags
}

// ConvertToMB parses the memory size string and returns the equivalent number of megabytes.
// Supports units: k/K (kilobytes), m/M (megabytes), g/G (gigabytes), t/T (terabytes)
// The 'b' suffix is optional and ignored.
// Note: As per ML documentation, values are rounded down to the nearest MB for consistency.
func (v MemorySize) ConvertToMB() (int64, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("Memory Size Parse error", "memory size string value is null"))
		return 0, diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("Memory Size Parse Error", "memory size string value is unknown"))
		return 0, diags
	}

	valueString := v.ValueString()
	matches := memoryPattern.FindStringSubmatch(valueString)
	if len(matches) != 3 {
		diags.Append(diag.NewErrorDiagnostic("Memory Size Parse Error",
			fmt.Sprintf("invalid memory size format: %s", valueString)))
		return 0, diags
	}

	// Parse the numeric part
	numStr := matches[1]
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Memory Size Parse Error",
			fmt.Sprintf("invalid number in memory size: %s", numStr)))
		return 0, diags
	}

	// Parse the unit part (if present) and calculate bytes
	unit := strings.ToLower(matches[2])
	var bytes int64

	switch unit {
	case "k":
		bytes = num * 1024
	case "m":
		bytes = num * 1024 * 1024
	case "g":
		bytes = num * 1024 * 1024 * 1024
	case "t":
		bytes = num * 1024 * 1024 * 1024 * 1024
	case "": // no unit = bytes
		bytes = num
	default:
		diags.Append(diag.NewErrorDiagnostic("Memory Size Parse Error",
			fmt.Sprintf("unsupported memory unit: %s", unit)))
		return 0, diags
	}

	// Round down to the nearest MB (1024*1024 = 1048576 bytes) as per ML documentation
	const mbInBytes = 1024 * 1024
	roundedMB := bytes / mbInBytes

	return roundedMB, diags
}

// NewMemorySizeNull creates a MemorySize with a null value. Determine whether the value is null via IsNull method.
func NewMemorySizeNull() MemorySize {
	return MemorySize{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewMemorySizeUnknown creates a MemorySize with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewMemorySizeUnknown() MemorySize {
	return MemorySize{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewMemorySizeValue creates a MemorySize with a known value. Access the value via ValueString method.
func NewMemorySizeValue(value string) MemorySize {
	return MemorySize{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewMemorySizePointerValue creates a MemorySize with a null value if nil or a known value. Access the value via ValueStringPointer method.
func NewMemorySizePointerValue(value *string) MemorySize {
	return MemorySize{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}

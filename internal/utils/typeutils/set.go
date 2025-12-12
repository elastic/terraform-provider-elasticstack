package typeutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NonEmptySetOrDefault[T any](ctx context.Context, original types.Set, elemType attr.Type, slice []T) (types.Set, diag.Diagnostics) {
	if len(slice) == 0 {
		return original, nil
	}

	return types.SetValueFrom(ctx, elemType, slice)
}

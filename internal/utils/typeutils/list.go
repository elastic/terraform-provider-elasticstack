package typeutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NonEmptyListOrDefault[T any](ctx context.Context, original types.List, elemType attr.Type, slice []T) (types.List, diag.Diagnostics) {
	if len(slice) == 0 {
		return original, nil
	}

	return types.ListValueFrom(ctx, elemType, slice)
}

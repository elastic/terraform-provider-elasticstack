package typeutils

import "github.com/hashicorp/terraform-plugin-framework/types"

// StringishPointerValue converts a pointer to a string-like type to a Terraform types.String value.
func StringishPointerValue[T ~string](ptr *T) types.String {
	if ptr == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*ptr))
}

// StringishValue converts a value of any string-like type T to a Terraform types.String.
func StringishValue[T ~string](value T) types.String {
	return types.StringValue(string(value))
}

func NonEmptyStringishValue[T ~string](value T) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(string(value))
}

func NonEmptyStringishPointerValue[T ~string](ptr *T) types.String {
	if ptr == nil {
		return types.StringNull()
	}
	return NonEmptyStringishValue(*ptr)
}

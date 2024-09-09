package kibana

func unwrapOptionalField[T any](field *T) T {
	var value T
	if field != nil {
		value = *field
	}

	return value
}

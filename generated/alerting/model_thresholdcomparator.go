/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
	"fmt"
)

// Thresholdcomparator The comparison function for the threshold. For example, \"is above\", \"is above or equals\", \"is below\", \"is below or equals\", \"is between\", and \"is not between\".
type Thresholdcomparator string

// List of thresholdcomparator
const (
	GREATER_THAN             Thresholdcomparator = ">"
	GREATER_THAN_OR_EQUAL_TO Thresholdcomparator = ">="
	LESS_THAN                Thresholdcomparator = "<"
	LESS_THAN_OR_EQUAL_TO    Thresholdcomparator = "<="
	BETWEEN                  Thresholdcomparator = "between"
	NOT_BETWEEN              Thresholdcomparator = "notBetween"
)

// All allowed values of Thresholdcomparator enum
var AllowedThresholdcomparatorEnumValues = []Thresholdcomparator{
	">",
	">=",
	"<",
	"<=",
	"between",
	"notBetween",
}

func (v *Thresholdcomparator) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := Thresholdcomparator(value)
	for _, existing := range AllowedThresholdcomparatorEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid Thresholdcomparator", value)
}

// NewThresholdcomparatorFromValue returns a pointer to a valid Thresholdcomparator
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewThresholdcomparatorFromValue(v string) (*Thresholdcomparator, error) {
	ev := Thresholdcomparator(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for Thresholdcomparator: valid values are %v", v, AllowedThresholdcomparatorEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v Thresholdcomparator) IsValid() bool {
	for _, existing := range AllowedThresholdcomparatorEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to thresholdcomparator value
func (v Thresholdcomparator) Ptr() *Thresholdcomparator {
	return &v
}

type NullableThresholdcomparator struct {
	value *Thresholdcomparator
	isSet bool
}

func (v NullableThresholdcomparator) Get() *Thresholdcomparator {
	return v.value
}

func (v *NullableThresholdcomparator) Set(val *Thresholdcomparator) {
	v.value = val
	v.isSet = true
}

func (v NullableThresholdcomparator) IsSet() bool {
	return v.isSet
}

func (v *NullableThresholdcomparator) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableThresholdcomparator(val *Thresholdcomparator) *NullableThresholdcomparator {
	return &NullableThresholdcomparator{value: val, isSet: true}
}

func (v NullableThresholdcomparator) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableThresholdcomparator) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

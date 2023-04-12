/*
Connectors

OpenAPI schema for Connectors endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package kibanaactions

import (
	"encoding/json"
	"fmt"
)

// RunConnectorRequestBodyPropertiesParams - struct for RunConnectorRequestBodyPropertiesParams
type RunConnectorRequestBodyPropertiesParams struct {
	RunConnectorParamsDocuments    *RunConnectorParamsDocuments
	RunConnectorParamsLevelMessage *RunConnectorParamsLevelMessage
	SubactionParameters            *SubactionParameters
}

// RunConnectorParamsDocumentsAsRunConnectorRequestBodyPropertiesParams is a convenience function that returns RunConnectorParamsDocuments wrapped in RunConnectorRequestBodyPropertiesParams
func RunConnectorParamsDocumentsAsRunConnectorRequestBodyPropertiesParams(v *RunConnectorParamsDocuments) RunConnectorRequestBodyPropertiesParams {
	return RunConnectorRequestBodyPropertiesParams{
		RunConnectorParamsDocuments: v,
	}
}

// RunConnectorParamsLevelMessageAsRunConnectorRequestBodyPropertiesParams is a convenience function that returns RunConnectorParamsLevelMessage wrapped in RunConnectorRequestBodyPropertiesParams
func RunConnectorParamsLevelMessageAsRunConnectorRequestBodyPropertiesParams(v *RunConnectorParamsLevelMessage) RunConnectorRequestBodyPropertiesParams {
	return RunConnectorRequestBodyPropertiesParams{
		RunConnectorParamsLevelMessage: v,
	}
}

// SubactionParametersAsRunConnectorRequestBodyPropertiesParams is a convenience function that returns SubactionParameters wrapped in RunConnectorRequestBodyPropertiesParams
func SubactionParametersAsRunConnectorRequestBodyPropertiesParams(v *SubactionParameters) RunConnectorRequestBodyPropertiesParams {
	return RunConnectorRequestBodyPropertiesParams{
		SubactionParameters: v,
	}
}

// Unmarshal JSON data into one of the pointers in the struct
func (dst *RunConnectorRequestBodyPropertiesParams) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into RunConnectorParamsDocuments
	err = newStrictDecoder(data).Decode(&dst.RunConnectorParamsDocuments)
	if err == nil {
		jsonRunConnectorParamsDocuments, _ := json.Marshal(dst.RunConnectorParamsDocuments)
		if string(jsonRunConnectorParamsDocuments) == "{}" { // empty struct
			dst.RunConnectorParamsDocuments = nil
		} else {
			match++
		}
	} else {
		dst.RunConnectorParamsDocuments = nil
	}

	// try to unmarshal data into RunConnectorParamsLevelMessage
	err = newStrictDecoder(data).Decode(&dst.RunConnectorParamsLevelMessage)
	if err == nil {
		jsonRunConnectorParamsLevelMessage, _ := json.Marshal(dst.RunConnectorParamsLevelMessage)
		if string(jsonRunConnectorParamsLevelMessage) == "{}" { // empty struct
			dst.RunConnectorParamsLevelMessage = nil
		} else {
			match++
		}
	} else {
		dst.RunConnectorParamsLevelMessage = nil
	}

	// try to unmarshal data into SubactionParameters
	err = newStrictDecoder(data).Decode(&dst.SubactionParameters)
	if err == nil {
		jsonSubactionParameters, _ := json.Marshal(dst.SubactionParameters)
		if string(jsonSubactionParameters) == "{}" { // empty struct
			dst.SubactionParameters = nil
		} else {
			match++
		}
	} else {
		dst.SubactionParameters = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		dst.RunConnectorParamsDocuments = nil
		dst.RunConnectorParamsLevelMessage = nil
		dst.SubactionParameters = nil

		return fmt.Errorf("data matches more than one schema in oneOf(RunConnectorRequestBodyPropertiesParams)")
	} else if match == 1 {
		return nil // exactly one match
	} else { // no match
		return fmt.Errorf("data failed to match schemas in oneOf(RunConnectorRequestBodyPropertiesParams)")
	}
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src RunConnectorRequestBodyPropertiesParams) MarshalJSON() ([]byte, error) {
	if src.RunConnectorParamsDocuments != nil {
		return json.Marshal(&src.RunConnectorParamsDocuments)
	}

	if src.RunConnectorParamsLevelMessage != nil {
		return json.Marshal(&src.RunConnectorParamsLevelMessage)
	}

	if src.SubactionParameters != nil {
		return json.Marshal(&src.SubactionParameters)
	}

	return nil, nil // no data in oneOf schemas
}

// Get the actual instance
func (obj *RunConnectorRequestBodyPropertiesParams) GetActualInstance() interface{} {
	if obj == nil {
		return nil
	}
	if obj.RunConnectorParamsDocuments != nil {
		return obj.RunConnectorParamsDocuments
	}

	if obj.RunConnectorParamsLevelMessage != nil {
		return obj.RunConnectorParamsLevelMessage
	}

	if obj.SubactionParameters != nil {
		return obj.SubactionParameters
	}

	// all schemas are nil
	return nil
}

type NullableRunConnectorRequestBodyPropertiesParams struct {
	value *RunConnectorRequestBodyPropertiesParams
	isSet bool
}

func (v NullableRunConnectorRequestBodyPropertiesParams) Get() *RunConnectorRequestBodyPropertiesParams {
	return v.value
}

func (v *NullableRunConnectorRequestBodyPropertiesParams) Set(val *RunConnectorRequestBodyPropertiesParams) {
	v.value = val
	v.isSet = true
}

func (v NullableRunConnectorRequestBodyPropertiesParams) IsSet() bool {
	return v.isSet
}

func (v *NullableRunConnectorRequestBodyPropertiesParams) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRunConnectorRequestBodyPropertiesParams(val *RunConnectorRequestBodyPropertiesParams) *NullableRunConnectorRequestBodyPropertiesParams {
	return &NullableRunConnectorRequestBodyPropertiesParams{value: val, isSet: true}
}

func (v NullableRunConnectorRequestBodyPropertiesParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRunConnectorRequestBodyPropertiesParams) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
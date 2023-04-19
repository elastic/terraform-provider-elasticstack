/*
SLOs

OpenAPI schema for SLOs endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package slo

import (
	"encoding/json"
	"fmt"
)

// SloResponseTimeWindow - struct for SloResponseTimeWindow
type SloResponseTimeWindow struct {
	TimeWindowCalendarAligned *TimeWindowCalendarAligned
	TimeWindowRolling *TimeWindowRolling
}

// TimeWindowCalendarAlignedAsSloResponseTimeWindow is a convenience function that returns TimeWindowCalendarAligned wrapped in SloResponseTimeWindow
func TimeWindowCalendarAlignedAsSloResponseTimeWindow(v *TimeWindowCalendarAligned) SloResponseTimeWindow {
	return SloResponseTimeWindow{
		TimeWindowCalendarAligned: v,
	}
}

// TimeWindowRollingAsSloResponseTimeWindow is a convenience function that returns TimeWindowRolling wrapped in SloResponseTimeWindow
func TimeWindowRollingAsSloResponseTimeWindow(v *TimeWindowRolling) SloResponseTimeWindow {
	return SloResponseTimeWindow{
		TimeWindowRolling: v,
	}
}


// Unmarshal JSON data into one of the pointers in the struct
func (dst *SloResponseTimeWindow) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into TimeWindowCalendarAligned
	err = newStrictDecoder(data).Decode(&dst.TimeWindowCalendarAligned)
	if err == nil {
		jsonTimeWindowCalendarAligned, _ := json.Marshal(dst.TimeWindowCalendarAligned)
		if string(jsonTimeWindowCalendarAligned) == "{}" { // empty struct
			dst.TimeWindowCalendarAligned = nil
		} else {
			match++
		}
	} else {
		dst.TimeWindowCalendarAligned = nil
	}

	// try to unmarshal data into TimeWindowRolling
	err = newStrictDecoder(data).Decode(&dst.TimeWindowRolling)
	if err == nil {
		jsonTimeWindowRolling, _ := json.Marshal(dst.TimeWindowRolling)
		if string(jsonTimeWindowRolling) == "{}" { // empty struct
			dst.TimeWindowRolling = nil
		} else {
			match++
		}
	} else {
		dst.TimeWindowRolling = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		dst.TimeWindowCalendarAligned = nil
		dst.TimeWindowRolling = nil

		return fmt.Errorf("data matches more than one schema in oneOf(SloResponseTimeWindow)")
	} else if match == 1 {
		return nil // exactly one match
	} else { // no match
		return fmt.Errorf("data failed to match schemas in oneOf(SloResponseTimeWindow)")
	}
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src SloResponseTimeWindow) MarshalJSON() ([]byte, error) {
	if src.TimeWindowCalendarAligned != nil {
		return json.Marshal(&src.TimeWindowCalendarAligned)
	}

	if src.TimeWindowRolling != nil {
		return json.Marshal(&src.TimeWindowRolling)
	}

	return nil, nil // no data in oneOf schemas
}

// Get the actual instance
func (obj *SloResponseTimeWindow) GetActualInstance() (interface{}) {
	if obj == nil {
		return nil
	}
	if obj.TimeWindowCalendarAligned != nil {
		return obj.TimeWindowCalendarAligned
	}

	if obj.TimeWindowRolling != nil {
		return obj.TimeWindowRolling
	}

	// all schemas are nil
	return nil
}

type NullableSloResponseTimeWindow struct {
	value *SloResponseTimeWindow
	isSet bool
}

func (v NullableSloResponseTimeWindow) Get() *SloResponseTimeWindow {
	return v.value
}

func (v *NullableSloResponseTimeWindow) Set(val *SloResponseTimeWindow) {
	v.value = val
	v.isSet = true
}

func (v NullableSloResponseTimeWindow) IsSet() bool {
	return v.isSet
}

func (v *NullableSloResponseTimeWindow) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSloResponseTimeWindow(val *SloResponseTimeWindow) *NullableSloResponseTimeWindow {
	return &NullableSloResponseTimeWindow{value: val, isSet: true}
}

func (v NullableSloResponseTimeWindow) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSloResponseTimeWindow) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



/*
SLOs

OpenAPI schema for SLOs endpoints

API version: 1.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package slo

import (
	"encoding/json"
)

// checks if the FindSloDefinitionsResponseOneOf1 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FindSloDefinitionsResponseOneOf1{}

// FindSloDefinitionsResponseOneOf1 struct for FindSloDefinitionsResponseOneOf1
type FindSloDefinitionsResponseOneOf1 struct {
	// for backward compability
	Page *float64 `json:"page,omitempty"`
	// for backward compability
	PerPage *float64 `json:"perPage,omitempty"`
	Size    *float64 `json:"size,omitempty"`
	// the cursor to provide to get the next paged results
	SearchAfter []string                 `json:"searchAfter,omitempty"`
	Total       *float64                 `json:"total,omitempty"`
	Results     []SloWithSummaryResponse `json:"results,omitempty"`
}

// NewFindSloDefinitionsResponseOneOf1 instantiates a new FindSloDefinitionsResponseOneOf1 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFindSloDefinitionsResponseOneOf1() *FindSloDefinitionsResponseOneOf1 {
	this := FindSloDefinitionsResponseOneOf1{}
	var page float64 = 1
	this.Page = &page
	return &this
}

// NewFindSloDefinitionsResponseOneOf1WithDefaults instantiates a new FindSloDefinitionsResponseOneOf1 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFindSloDefinitionsResponseOneOf1WithDefaults() *FindSloDefinitionsResponseOneOf1 {
	this := FindSloDefinitionsResponseOneOf1{}
	var page float64 = 1
	this.Page = &page
	return &this
}

// GetPage returns the Page field value if set, zero value otherwise.
func (o *FindSloDefinitionsResponseOneOf1) GetPage() float64 {
	if o == nil || IsNil(o.Page) {
		var ret float64
		return ret
	}
	return *o.Page
}

// GetPageOk returns a tuple with the Page field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FindSloDefinitionsResponseOneOf1) GetPageOk() (*float64, bool) {
	if o == nil || IsNil(o.Page) {
		return nil, false
	}
	return o.Page, true
}

// HasPage returns a boolean if a field has been set.
func (o *FindSloDefinitionsResponseOneOf1) HasPage() bool {
	if o != nil && !IsNil(o.Page) {
		return true
	}

	return false
}

// SetPage gets a reference to the given float64 and assigns it to the Page field.
func (o *FindSloDefinitionsResponseOneOf1) SetPage(v float64) {
	o.Page = &v
}

// GetPerPage returns the PerPage field value if set, zero value otherwise.
func (o *FindSloDefinitionsResponseOneOf1) GetPerPage() float64 {
	if o == nil || IsNil(o.PerPage) {
		var ret float64
		return ret
	}
	return *o.PerPage
}

// GetPerPageOk returns a tuple with the PerPage field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FindSloDefinitionsResponseOneOf1) GetPerPageOk() (*float64, bool) {
	if o == nil || IsNil(o.PerPage) {
		return nil, false
	}
	return o.PerPage, true
}

// HasPerPage returns a boolean if a field has been set.
func (o *FindSloDefinitionsResponseOneOf1) HasPerPage() bool {
	if o != nil && !IsNil(o.PerPage) {
		return true
	}

	return false
}

// SetPerPage gets a reference to the given float64 and assigns it to the PerPage field.
func (o *FindSloDefinitionsResponseOneOf1) SetPerPage(v float64) {
	o.PerPage = &v
}

// GetSize returns the Size field value if set, zero value otherwise.
func (o *FindSloDefinitionsResponseOneOf1) GetSize() float64 {
	if o == nil || IsNil(o.Size) {
		var ret float64
		return ret
	}
	return *o.Size
}

// GetSizeOk returns a tuple with the Size field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FindSloDefinitionsResponseOneOf1) GetSizeOk() (*float64, bool) {
	if o == nil || IsNil(o.Size) {
		return nil, false
	}
	return o.Size, true
}

// HasSize returns a boolean if a field has been set.
func (o *FindSloDefinitionsResponseOneOf1) HasSize() bool {
	if o != nil && !IsNil(o.Size) {
		return true
	}

	return false
}

// SetSize gets a reference to the given float64 and assigns it to the Size field.
func (o *FindSloDefinitionsResponseOneOf1) SetSize(v float64) {
	o.Size = &v
}

// GetSearchAfter returns the SearchAfter field value if set, zero value otherwise.
func (o *FindSloDefinitionsResponseOneOf1) GetSearchAfter() []string {
	if o == nil || IsNil(o.SearchAfter) {
		var ret []string
		return ret
	}
	return o.SearchAfter
}

// GetSearchAfterOk returns a tuple with the SearchAfter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FindSloDefinitionsResponseOneOf1) GetSearchAfterOk() ([]string, bool) {
	if o == nil || IsNil(o.SearchAfter) {
		return nil, false
	}
	return o.SearchAfter, true
}

// HasSearchAfter returns a boolean if a field has been set.
func (o *FindSloDefinitionsResponseOneOf1) HasSearchAfter() bool {
	if o != nil && !IsNil(o.SearchAfter) {
		return true
	}

	return false
}

// SetSearchAfter gets a reference to the given []string and assigns it to the SearchAfter field.
func (o *FindSloDefinitionsResponseOneOf1) SetSearchAfter(v []string) {
	o.SearchAfter = v
}

// GetTotal returns the Total field value if set, zero value otherwise.
func (o *FindSloDefinitionsResponseOneOf1) GetTotal() float64 {
	if o == nil || IsNil(o.Total) {
		var ret float64
		return ret
	}
	return *o.Total
}

// GetTotalOk returns a tuple with the Total field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FindSloDefinitionsResponseOneOf1) GetTotalOk() (*float64, bool) {
	if o == nil || IsNil(o.Total) {
		return nil, false
	}
	return o.Total, true
}

// HasTotal returns a boolean if a field has been set.
func (o *FindSloDefinitionsResponseOneOf1) HasTotal() bool {
	if o != nil && !IsNil(o.Total) {
		return true
	}

	return false
}

// SetTotal gets a reference to the given float64 and assigns it to the Total field.
func (o *FindSloDefinitionsResponseOneOf1) SetTotal(v float64) {
	o.Total = &v
}

// GetResults returns the Results field value if set, zero value otherwise.
func (o *FindSloDefinitionsResponseOneOf1) GetResults() []SloWithSummaryResponse {
	if o == nil || IsNil(o.Results) {
		var ret []SloWithSummaryResponse
		return ret
	}
	return o.Results
}

// GetResultsOk returns a tuple with the Results field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *FindSloDefinitionsResponseOneOf1) GetResultsOk() ([]SloWithSummaryResponse, bool) {
	if o == nil || IsNil(o.Results) {
		return nil, false
	}
	return o.Results, true
}

// HasResults returns a boolean if a field has been set.
func (o *FindSloDefinitionsResponseOneOf1) HasResults() bool {
	if o != nil && !IsNil(o.Results) {
		return true
	}

	return false
}

// SetResults gets a reference to the given []SloWithSummaryResponse and assigns it to the Results field.
func (o *FindSloDefinitionsResponseOneOf1) SetResults(v []SloWithSummaryResponse) {
	o.Results = v
}

func (o FindSloDefinitionsResponseOneOf1) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FindSloDefinitionsResponseOneOf1) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Page) {
		toSerialize["page"] = o.Page
	}
	if !IsNil(o.PerPage) {
		toSerialize["perPage"] = o.PerPage
	}
	if !IsNil(o.Size) {
		toSerialize["size"] = o.Size
	}
	if !IsNil(o.SearchAfter) {
		toSerialize["searchAfter"] = o.SearchAfter
	}
	if !IsNil(o.Total) {
		toSerialize["total"] = o.Total
	}
	if !IsNil(o.Results) {
		toSerialize["results"] = o.Results
	}
	return toSerialize, nil
}

type NullableFindSloDefinitionsResponseOneOf1 struct {
	value *FindSloDefinitionsResponseOneOf1
	isSet bool
}

func (v NullableFindSloDefinitionsResponseOneOf1) Get() *FindSloDefinitionsResponseOneOf1 {
	return v.value
}

func (v *NullableFindSloDefinitionsResponseOneOf1) Set(val *FindSloDefinitionsResponseOneOf1) {
	v.value = val
	v.isSet = true
}

func (v NullableFindSloDefinitionsResponseOneOf1) IsSet() bool {
	return v.isSet
}

func (v *NullableFindSloDefinitionsResponseOneOf1) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFindSloDefinitionsResponseOneOf1(val *FindSloDefinitionsResponseOneOf1) *NullableFindSloDefinitionsResponseOneOf1 {
	return &NullableFindSloDefinitionsResponseOneOf1{value: val, isSet: true}
}

func (v NullableFindSloDefinitionsResponseOneOf1) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFindSloDefinitionsResponseOneOf1) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
)

// checks if the ParamsEsQueryRuleOneOf1 type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ParamsEsQueryRuleOneOf1{}

// ParamsEsQueryRuleOneOf1 The parameters for an Elasticsearch query rule that uses KQL or Lucene to define the query.
type ParamsEsQueryRuleOneOf1 struct {
	// The name of the numeric field that is used in the aggregation. This property is required when `aggType` is `avg`, `max`, `min` or `sum`.
	AggField *string  `json:"aggField,omitempty"`
	AggType  *Aggtype `json:"aggType,omitempty"`
	// Indicates whether to exclude matches from previous runs. If `true`, you can avoid alert duplication by excluding documents that have already been detected by the previous rule run. This option is not available when a grouping field is specified.
	ExcludeHitsFromPreviousRun *bool                                       `json:"excludeHitsFromPreviousRun,omitempty"`
	GroupBy                    *Groupby                                    `json:"groupBy,omitempty"`
	SearchConfiguration        *ParamsEsQueryRuleOneOf1SearchConfiguration `json:"searchConfiguration,omitempty"`
	// The type of query, in this case a text-based query that uses KQL or Lucene.
	SearchType string `json:"searchType"`
	// The number of documents to pass to the configured actions when the threshold condition is met.
	Size      int32      `json:"size"`
	TermField *Termfield `json:"termField,omitempty"`
	// This property is required when `groupBy` is `top`. It specifies the number of groups to check against the threshold and therefore limits the number of alerts on high cardinality fields.
	TermSize *int32 `json:"termSize,omitempty"`
	// The threshold value that is used with the `thresholdComparator`. If the `thresholdComparator` is `between` or `notBetween`, you must specify the boundary values.
	Threshold           []int32             `json:"threshold"`
	ThresholdComparator Thresholdcomparator `json:"thresholdComparator"`
	// The field that is used to calculate the time window.
	TimeField *string `json:"timeField,omitempty"`
	// The size of the time window (in `timeWindowUnit` units), which determines how far back to search for documents. Generally it should be a value higher than the rule check interval to avoid gaps in detection.
	TimeWindowSize int32          `json:"timeWindowSize"`
	TimeWindowUnit Timewindowunit `json:"timeWindowUnit"`
}

// NewParamsEsQueryRuleOneOf1 instantiates a new ParamsEsQueryRuleOneOf1 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewParamsEsQueryRuleOneOf1(searchType string, size int32, threshold []int32, thresholdComparator Thresholdcomparator, timeWindowSize int32, timeWindowUnit Timewindowunit) *ParamsEsQueryRuleOneOf1 {
	this := ParamsEsQueryRuleOneOf1{}
	var aggType Aggtype = COUNT
	this.AggType = &aggType
	var groupBy Groupby = ALL
	this.GroupBy = &groupBy
	this.SearchType = searchType
	this.Size = size
	this.Threshold = threshold
	this.ThresholdComparator = thresholdComparator
	this.TimeWindowSize = timeWindowSize
	this.TimeWindowUnit = timeWindowUnit
	return &this
}

// NewParamsEsQueryRuleOneOf1WithDefaults instantiates a new ParamsEsQueryRuleOneOf1 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewParamsEsQueryRuleOneOf1WithDefaults() *ParamsEsQueryRuleOneOf1 {
	this := ParamsEsQueryRuleOneOf1{}
	var aggType Aggtype = COUNT
	this.AggType = &aggType
	var groupBy Groupby = ALL
	this.GroupBy = &groupBy
	return &this
}

// GetAggField returns the AggField field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetAggField() string {
	if o == nil || IsNil(o.AggField) {
		var ret string
		return ret
	}
	return *o.AggField
}

// GetAggFieldOk returns a tuple with the AggField field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetAggFieldOk() (*string, bool) {
	if o == nil || IsNil(o.AggField) {
		return nil, false
	}
	return o.AggField, true
}

// HasAggField returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasAggField() bool {
	if o != nil && !IsNil(o.AggField) {
		return true
	}

	return false
}

// SetAggField gets a reference to the given string and assigns it to the AggField field.
func (o *ParamsEsQueryRuleOneOf1) SetAggField(v string) {
	o.AggField = &v
}

// GetAggType returns the AggType field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetAggType() Aggtype {
	if o == nil || IsNil(o.AggType) {
		var ret Aggtype
		return ret
	}
	return *o.AggType
}

// GetAggTypeOk returns a tuple with the AggType field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetAggTypeOk() (*Aggtype, bool) {
	if o == nil || IsNil(o.AggType) {
		return nil, false
	}
	return o.AggType, true
}

// HasAggType returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasAggType() bool {
	if o != nil && !IsNil(o.AggType) {
		return true
	}

	return false
}

// SetAggType gets a reference to the given Aggtype and assigns it to the AggType field.
func (o *ParamsEsQueryRuleOneOf1) SetAggType(v Aggtype) {
	o.AggType = &v
}

// GetExcludeHitsFromPreviousRun returns the ExcludeHitsFromPreviousRun field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetExcludeHitsFromPreviousRun() bool {
	if o == nil || IsNil(o.ExcludeHitsFromPreviousRun) {
		var ret bool
		return ret
	}
	return *o.ExcludeHitsFromPreviousRun
}

// GetExcludeHitsFromPreviousRunOk returns a tuple with the ExcludeHitsFromPreviousRun field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetExcludeHitsFromPreviousRunOk() (*bool, bool) {
	if o == nil || IsNil(o.ExcludeHitsFromPreviousRun) {
		return nil, false
	}
	return o.ExcludeHitsFromPreviousRun, true
}

// HasExcludeHitsFromPreviousRun returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasExcludeHitsFromPreviousRun() bool {
	if o != nil && !IsNil(o.ExcludeHitsFromPreviousRun) {
		return true
	}

	return false
}

// SetExcludeHitsFromPreviousRun gets a reference to the given bool and assigns it to the ExcludeHitsFromPreviousRun field.
func (o *ParamsEsQueryRuleOneOf1) SetExcludeHitsFromPreviousRun(v bool) {
	o.ExcludeHitsFromPreviousRun = &v
}

// GetGroupBy returns the GroupBy field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetGroupBy() Groupby {
	if o == nil || IsNil(o.GroupBy) {
		var ret Groupby
		return ret
	}
	return *o.GroupBy
}

// GetGroupByOk returns a tuple with the GroupBy field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetGroupByOk() (*Groupby, bool) {
	if o == nil || IsNil(o.GroupBy) {
		return nil, false
	}
	return o.GroupBy, true
}

// HasGroupBy returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasGroupBy() bool {
	if o != nil && !IsNil(o.GroupBy) {
		return true
	}

	return false
}

// SetGroupBy gets a reference to the given Groupby and assigns it to the GroupBy field.
func (o *ParamsEsQueryRuleOneOf1) SetGroupBy(v Groupby) {
	o.GroupBy = &v
}

// GetSearchConfiguration returns the SearchConfiguration field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetSearchConfiguration() ParamsEsQueryRuleOneOf1SearchConfiguration {
	if o == nil || IsNil(o.SearchConfiguration) {
		var ret ParamsEsQueryRuleOneOf1SearchConfiguration
		return ret
	}
	return *o.SearchConfiguration
}

// GetSearchConfigurationOk returns a tuple with the SearchConfiguration field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetSearchConfigurationOk() (*ParamsEsQueryRuleOneOf1SearchConfiguration, bool) {
	if o == nil || IsNil(o.SearchConfiguration) {
		return nil, false
	}
	return o.SearchConfiguration, true
}

// HasSearchConfiguration returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasSearchConfiguration() bool {
	if o != nil && !IsNil(o.SearchConfiguration) {
		return true
	}

	return false
}

// SetSearchConfiguration gets a reference to the given ParamsEsQueryRuleOneOf1SearchConfiguration and assigns it to the SearchConfiguration field.
func (o *ParamsEsQueryRuleOneOf1) SetSearchConfiguration(v ParamsEsQueryRuleOneOf1SearchConfiguration) {
	o.SearchConfiguration = &v
}

// GetSearchType returns the SearchType field value
func (o *ParamsEsQueryRuleOneOf1) GetSearchType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.SearchType
}

// GetSearchTypeOk returns a tuple with the SearchType field value
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetSearchTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.SearchType, true
}

// SetSearchType sets field value
func (o *ParamsEsQueryRuleOneOf1) SetSearchType(v string) {
	o.SearchType = v
}

// GetSize returns the Size field value
func (o *ParamsEsQueryRuleOneOf1) GetSize() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Size
}

// GetSizeOk returns a tuple with the Size field value
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetSizeOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Size, true
}

// SetSize sets field value
func (o *ParamsEsQueryRuleOneOf1) SetSize(v int32) {
	o.Size = v
}

// GetTermField returns the TermField field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetTermField() Termfield {
	if o == nil || IsNil(o.TermField) {
		var ret Termfield
		return ret
	}
	return *o.TermField
}

// GetTermFieldOk returns a tuple with the TermField field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetTermFieldOk() (*Termfield, bool) {
	if o == nil || IsNil(o.TermField) {
		return nil, false
	}
	return o.TermField, true
}

// HasTermField returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasTermField() bool {
	if o != nil && !IsNil(o.TermField) {
		return true
	}

	return false
}

// SetTermField gets a reference to the given Termfield and assigns it to the TermField field.
func (o *ParamsEsQueryRuleOneOf1) SetTermField(v Termfield) {
	o.TermField = &v
}

// GetTermSize returns the TermSize field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetTermSize() int32 {
	if o == nil || IsNil(o.TermSize) {
		var ret int32
		return ret
	}
	return *o.TermSize
}

// GetTermSizeOk returns a tuple with the TermSize field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetTermSizeOk() (*int32, bool) {
	if o == nil || IsNil(o.TermSize) {
		return nil, false
	}
	return o.TermSize, true
}

// HasTermSize returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasTermSize() bool {
	if o != nil && !IsNil(o.TermSize) {
		return true
	}

	return false
}

// SetTermSize gets a reference to the given int32 and assigns it to the TermSize field.
func (o *ParamsEsQueryRuleOneOf1) SetTermSize(v int32) {
	o.TermSize = &v
}

// GetThreshold returns the Threshold field value
func (o *ParamsEsQueryRuleOneOf1) GetThreshold() []int32 {
	if o == nil {
		var ret []int32
		return ret
	}

	return o.Threshold
}

// GetThresholdOk returns a tuple with the Threshold field value
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetThresholdOk() ([]int32, bool) {
	if o == nil {
		return nil, false
	}
	return o.Threshold, true
}

// SetThreshold sets field value
func (o *ParamsEsQueryRuleOneOf1) SetThreshold(v []int32) {
	o.Threshold = v
}

// GetThresholdComparator returns the ThresholdComparator field value
func (o *ParamsEsQueryRuleOneOf1) GetThresholdComparator() Thresholdcomparator {
	if o == nil {
		var ret Thresholdcomparator
		return ret
	}

	return o.ThresholdComparator
}

// GetThresholdComparatorOk returns a tuple with the ThresholdComparator field value
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetThresholdComparatorOk() (*Thresholdcomparator, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ThresholdComparator, true
}

// SetThresholdComparator sets field value
func (o *ParamsEsQueryRuleOneOf1) SetThresholdComparator(v Thresholdcomparator) {
	o.ThresholdComparator = v
}

// GetTimeField returns the TimeField field value if set, zero value otherwise.
func (o *ParamsEsQueryRuleOneOf1) GetTimeField() string {
	if o == nil || IsNil(o.TimeField) {
		var ret string
		return ret
	}
	return *o.TimeField
}

// GetTimeFieldOk returns a tuple with the TimeField field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetTimeFieldOk() (*string, bool) {
	if o == nil || IsNil(o.TimeField) {
		return nil, false
	}
	return o.TimeField, true
}

// HasTimeField returns a boolean if a field has been set.
func (o *ParamsEsQueryRuleOneOf1) HasTimeField() bool {
	if o != nil && !IsNil(o.TimeField) {
		return true
	}

	return false
}

// SetTimeField gets a reference to the given string and assigns it to the TimeField field.
func (o *ParamsEsQueryRuleOneOf1) SetTimeField(v string) {
	o.TimeField = &v
}

// GetTimeWindowSize returns the TimeWindowSize field value
func (o *ParamsEsQueryRuleOneOf1) GetTimeWindowSize() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.TimeWindowSize
}

// GetTimeWindowSizeOk returns a tuple with the TimeWindowSize field value
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetTimeWindowSizeOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimeWindowSize, true
}

// SetTimeWindowSize sets field value
func (o *ParamsEsQueryRuleOneOf1) SetTimeWindowSize(v int32) {
	o.TimeWindowSize = v
}

// GetTimeWindowUnit returns the TimeWindowUnit field value
func (o *ParamsEsQueryRuleOneOf1) GetTimeWindowUnit() Timewindowunit {
	if o == nil {
		var ret Timewindowunit
		return ret
	}

	return o.TimeWindowUnit
}

// GetTimeWindowUnitOk returns a tuple with the TimeWindowUnit field value
// and a boolean to check if the value has been set.
func (o *ParamsEsQueryRuleOneOf1) GetTimeWindowUnitOk() (*Timewindowunit, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimeWindowUnit, true
}

// SetTimeWindowUnit sets field value
func (o *ParamsEsQueryRuleOneOf1) SetTimeWindowUnit(v Timewindowunit) {
	o.TimeWindowUnit = v
}

func (o ParamsEsQueryRuleOneOf1) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ParamsEsQueryRuleOneOf1) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.AggField) {
		toSerialize["aggField"] = o.AggField
	}
	if !IsNil(o.AggType) {
		toSerialize["aggType"] = o.AggType
	}
	if !IsNil(o.ExcludeHitsFromPreviousRun) {
		toSerialize["excludeHitsFromPreviousRun"] = o.ExcludeHitsFromPreviousRun
	}
	if !IsNil(o.GroupBy) {
		toSerialize["groupBy"] = o.GroupBy
	}
	if !IsNil(o.SearchConfiguration) {
		toSerialize["searchConfiguration"] = o.SearchConfiguration
	}
	toSerialize["searchType"] = o.SearchType
	toSerialize["size"] = o.Size
	if !IsNil(o.TermField) {
		toSerialize["termField"] = o.TermField
	}
	if !IsNil(o.TermSize) {
		toSerialize["termSize"] = o.TermSize
	}
	toSerialize["threshold"] = o.Threshold
	toSerialize["thresholdComparator"] = o.ThresholdComparator
	if !IsNil(o.TimeField) {
		toSerialize["timeField"] = o.TimeField
	}
	toSerialize["timeWindowSize"] = o.TimeWindowSize
	toSerialize["timeWindowUnit"] = o.TimeWindowUnit
	return toSerialize, nil
}

type NullableParamsEsQueryRuleOneOf1 struct {
	value *ParamsEsQueryRuleOneOf1
	isSet bool
}

func (v NullableParamsEsQueryRuleOneOf1) Get() *ParamsEsQueryRuleOneOf1 {
	return v.value
}

func (v *NullableParamsEsQueryRuleOneOf1) Set(val *ParamsEsQueryRuleOneOf1) {
	v.value = val
	v.isSet = true
}

func (v NullableParamsEsQueryRuleOneOf1) IsSet() bool {
	return v.isSet
}

func (v *NullableParamsEsQueryRuleOneOf1) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableParamsEsQueryRuleOneOf1(val *ParamsEsQueryRuleOneOf1) *NullableParamsEsQueryRuleOneOf1 {
	return &NullableParamsEsQueryRuleOneOf1{value: val, isSet: true}
}

func (v NullableParamsEsQueryRuleOneOf1) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableParamsEsQueryRuleOneOf1) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

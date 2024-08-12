# ParamsIndexThresholdRule

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AggField** | Pointer to **string** | The name of the numeric field that is used in the aggregation. This property is required when &#x60;aggType&#x60; is &#x60;avg&#x60;, &#x60;max&#x60;, &#x60;min&#x60; or &#x60;sum&#x60;.  | [optional] 
**AggType** | Pointer to [**Aggtype**](Aggtype.md) |  | [optional] [default to COUNT]
**FilterKuery** | Pointer to **string** | A KQL expression thats limits the scope of alerts. | [optional] 
**GroupBy** | Pointer to [**Groupby**](Groupby.md) |  | [optional] [default to ALL]
**Index** | **[]string** |  | 
**TermField** | Pointer to **NullableString** |  | [optional] 
**TermSize** | Pointer to **int32** | This property is required when &#x60;groupBy&#x60; is &#x60;top&#x60;. It specifies the number of groups to check against the threshold and therefore limits the number of alerts on high cardinality fields.  | [optional] 
**Threshold** | **[]int32** |  | 
**ThresholdComparator** | [**Thresholdcomparator**](Thresholdcomparator.md) |  | 
**TimeField** | **string** | The field that is used to calculate the time window. | 
**TimeWindowSize** | **int32** | The size of the time window (in &#x60;timeWindowUnit&#x60; units), which determines how far back to search for documents. Generally it should be a value higher than the rule check interval to avoid gaps in detection.  | 
**TimeWindowUnit** | [**Timewindowunit**](Timewindowunit.md) |  | 

## Methods

### NewParamsIndexThresholdRule

`func NewParamsIndexThresholdRule(index []string, threshold []int32, thresholdComparator Thresholdcomparator, timeField string, timeWindowSize int32, timeWindowUnit Timewindowunit, ) *ParamsIndexThresholdRule`

NewParamsIndexThresholdRule instantiates a new ParamsIndexThresholdRule object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsIndexThresholdRuleWithDefaults

`func NewParamsIndexThresholdRuleWithDefaults() *ParamsIndexThresholdRule`

NewParamsIndexThresholdRuleWithDefaults instantiates a new ParamsIndexThresholdRule object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAggField

`func (o *ParamsIndexThresholdRule) GetAggField() string`

GetAggField returns the AggField field if non-nil, zero value otherwise.

### GetAggFieldOk

`func (o *ParamsIndexThresholdRule) GetAggFieldOk() (*string, bool)`

GetAggFieldOk returns a tuple with the AggField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggField

`func (o *ParamsIndexThresholdRule) SetAggField(v string)`

SetAggField sets AggField field to given value.

### HasAggField

`func (o *ParamsIndexThresholdRule) HasAggField() bool`

HasAggField returns a boolean if a field has been set.

### GetAggType

`func (o *ParamsIndexThresholdRule) GetAggType() Aggtype`

GetAggType returns the AggType field if non-nil, zero value otherwise.

### GetAggTypeOk

`func (o *ParamsIndexThresholdRule) GetAggTypeOk() (*Aggtype, bool)`

GetAggTypeOk returns a tuple with the AggType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggType

`func (o *ParamsIndexThresholdRule) SetAggType(v Aggtype)`

SetAggType sets AggType field to given value.

### HasAggType

`func (o *ParamsIndexThresholdRule) HasAggType() bool`

HasAggType returns a boolean if a field has been set.

### GetFilterKuery

`func (o *ParamsIndexThresholdRule) GetFilterKuery() string`

GetFilterKuery returns the FilterKuery field if non-nil, zero value otherwise.

### GetFilterKueryOk

`func (o *ParamsIndexThresholdRule) GetFilterKueryOk() (*string, bool)`

GetFilterKueryOk returns a tuple with the FilterKuery field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilterKuery

`func (o *ParamsIndexThresholdRule) SetFilterKuery(v string)`

SetFilterKuery sets FilterKuery field to given value.

### HasFilterKuery

`func (o *ParamsIndexThresholdRule) HasFilterKuery() bool`

HasFilterKuery returns a boolean if a field has been set.

### GetGroupBy

`func (o *ParamsIndexThresholdRule) GetGroupBy() Groupby`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *ParamsIndexThresholdRule) GetGroupByOk() (*Groupby, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *ParamsIndexThresholdRule) SetGroupBy(v Groupby)`

SetGroupBy sets GroupBy field to given value.

### HasGroupBy

`func (o *ParamsIndexThresholdRule) HasGroupBy() bool`

HasGroupBy returns a boolean if a field has been set.

### GetIndex

`func (o *ParamsIndexThresholdRule) GetIndex() []string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *ParamsIndexThresholdRule) GetIndexOk() (*[]string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *ParamsIndexThresholdRule) SetIndex(v []string)`

SetIndex sets Index field to given value.


### GetTermField

`func (o *ParamsIndexThresholdRule) GetTermField() string`

GetTermField returns the TermField field if non-nil, zero value otherwise.

### GetTermFieldOk

`func (o *ParamsIndexThresholdRule) GetTermFieldOk() (*string, bool)`

GetTermFieldOk returns a tuple with the TermField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTermField

`func (o *ParamsIndexThresholdRule) SetTermField(v string)`

SetTermField sets TermField field to given value.

### HasTermField

`func (o *ParamsIndexThresholdRule) HasTermField() bool`

HasTermField returns a boolean if a field has been set.

### SetTermFieldNil

`func (o *ParamsIndexThresholdRule) SetTermFieldNil(b bool)`

 SetTermFieldNil sets the value for TermField to be an explicit nil

### UnsetTermField
`func (o *ParamsIndexThresholdRule) UnsetTermField()`

UnsetTermField ensures that no value is present for TermField, not even an explicit nil
### GetTermSize

`func (o *ParamsIndexThresholdRule) GetTermSize() int32`

GetTermSize returns the TermSize field if non-nil, zero value otherwise.

### GetTermSizeOk

`func (o *ParamsIndexThresholdRule) GetTermSizeOk() (*int32, bool)`

GetTermSizeOk returns a tuple with the TermSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTermSize

`func (o *ParamsIndexThresholdRule) SetTermSize(v int32)`

SetTermSize sets TermSize field to given value.

### HasTermSize

`func (o *ParamsIndexThresholdRule) HasTermSize() bool`

HasTermSize returns a boolean if a field has been set.

### GetThreshold

`func (o *ParamsIndexThresholdRule) GetThreshold() []int32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *ParamsIndexThresholdRule) GetThresholdOk() (*[]int32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *ParamsIndexThresholdRule) SetThreshold(v []int32)`

SetThreshold sets Threshold field to given value.


### GetThresholdComparator

`func (o *ParamsIndexThresholdRule) GetThresholdComparator() Thresholdcomparator`

GetThresholdComparator returns the ThresholdComparator field if non-nil, zero value otherwise.

### GetThresholdComparatorOk

`func (o *ParamsIndexThresholdRule) GetThresholdComparatorOk() (*Thresholdcomparator, bool)`

GetThresholdComparatorOk returns a tuple with the ThresholdComparator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThresholdComparator

`func (o *ParamsIndexThresholdRule) SetThresholdComparator(v Thresholdcomparator)`

SetThresholdComparator sets ThresholdComparator field to given value.


### GetTimeField

`func (o *ParamsIndexThresholdRule) GetTimeField() string`

GetTimeField returns the TimeField field if non-nil, zero value otherwise.

### GetTimeFieldOk

`func (o *ParamsIndexThresholdRule) GetTimeFieldOk() (*string, bool)`

GetTimeFieldOk returns a tuple with the TimeField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeField

`func (o *ParamsIndexThresholdRule) SetTimeField(v string)`

SetTimeField sets TimeField field to given value.


### GetTimeWindowSize

`func (o *ParamsIndexThresholdRule) GetTimeWindowSize() int32`

GetTimeWindowSize returns the TimeWindowSize field if non-nil, zero value otherwise.

### GetTimeWindowSizeOk

`func (o *ParamsIndexThresholdRule) GetTimeWindowSizeOk() (*int32, bool)`

GetTimeWindowSizeOk returns a tuple with the TimeWindowSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindowSize

`func (o *ParamsIndexThresholdRule) SetTimeWindowSize(v int32)`

SetTimeWindowSize sets TimeWindowSize field to given value.


### GetTimeWindowUnit

`func (o *ParamsIndexThresholdRule) GetTimeWindowUnit() Timewindowunit`

GetTimeWindowUnit returns the TimeWindowUnit field if non-nil, zero value otherwise.

### GetTimeWindowUnitOk

`func (o *ParamsIndexThresholdRule) GetTimeWindowUnitOk() (*Timewindowunit, bool)`

GetTimeWindowUnitOk returns a tuple with the TimeWindowUnit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindowUnit

`func (o *ParamsIndexThresholdRule) SetTimeWindowUnit(v Timewindowunit)`

SetTimeWindowUnit sets TimeWindowUnit field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



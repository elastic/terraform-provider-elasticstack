# IndicatorPropertiesHistogramParamsTotal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | **string** | The field use to aggregate the good events. | 
**Aggregation** | **string** | The type of aggregation to use. | 
**Filter** | Pointer to **string** | The filter for total events. | [optional] 
**From** | Pointer to **float64** | The starting value of the range. Only required for \&quot;range\&quot; aggregations. | [optional] 
**To** | Pointer to **float64** | The ending value of the range. Only required for \&quot;range\&quot; aggregations. | [optional] 

## Methods

### NewIndicatorPropertiesHistogramParamsTotal

`func NewIndicatorPropertiesHistogramParamsTotal(field string, aggregation string, ) *IndicatorPropertiesHistogramParamsTotal`

NewIndicatorPropertiesHistogramParamsTotal instantiates a new IndicatorPropertiesHistogramParamsTotal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesHistogramParamsTotalWithDefaults

`func NewIndicatorPropertiesHistogramParamsTotalWithDefaults() *IndicatorPropertiesHistogramParamsTotal`

NewIndicatorPropertiesHistogramParamsTotalWithDefaults instantiates a new IndicatorPropertiesHistogramParamsTotal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *IndicatorPropertiesHistogramParamsTotal) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *IndicatorPropertiesHistogramParamsTotal) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *IndicatorPropertiesHistogramParamsTotal) SetField(v string)`

SetField sets Field field to given value.


### GetAggregation

`func (o *IndicatorPropertiesHistogramParamsTotal) GetAggregation() string`

GetAggregation returns the Aggregation field if non-nil, zero value otherwise.

### GetAggregationOk

`func (o *IndicatorPropertiesHistogramParamsTotal) GetAggregationOk() (*string, bool)`

GetAggregationOk returns a tuple with the Aggregation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggregation

`func (o *IndicatorPropertiesHistogramParamsTotal) SetAggregation(v string)`

SetAggregation sets Aggregation field to given value.


### GetFilter

`func (o *IndicatorPropertiesHistogramParamsTotal) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesHistogramParamsTotal) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesHistogramParamsTotal) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesHistogramParamsTotal) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetFrom

`func (o *IndicatorPropertiesHistogramParamsTotal) GetFrom() float64`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *IndicatorPropertiesHistogramParamsTotal) GetFromOk() (*float64, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *IndicatorPropertiesHistogramParamsTotal) SetFrom(v float64)`

SetFrom sets From field to given value.

### HasFrom

`func (o *IndicatorPropertiesHistogramParamsTotal) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetTo

`func (o *IndicatorPropertiesHistogramParamsTotal) GetTo() float64`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *IndicatorPropertiesHistogramParamsTotal) GetToOk() (*float64, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *IndicatorPropertiesHistogramParamsTotal) SetTo(v float64)`

SetTo sets To field to given value.

### HasTo

`func (o *IndicatorPropertiesHistogramParamsTotal) HasTo() bool`

HasTo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



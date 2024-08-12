# ParamsPropertyInfraInventoryCriteriaInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metric** | Pointer to **string** |  | [optional] 
**TimeSize** | Pointer to **float32** |  | [optional] 
**TimeUnit** | Pointer to **string** |  | [optional] 
**SourceId** | Pointer to **string** |  | [optional] 
**Threshold** | Pointer to **[]float32** |  | [optional] 
**Comparator** | Pointer to **string** |  | [optional] 
**CustomMetric** | Pointer to [**ParamsPropertyInfraInventoryCriteriaInnerCustomMetric**](ParamsPropertyInfraInventoryCriteriaInnerCustomMetric.md) |  | [optional] 
**WarningThreshold** | Pointer to **[]float32** |  | [optional] 
**WarningComparator** | Pointer to **string** |  | [optional] 

## Methods

### NewParamsPropertyInfraInventoryCriteriaInner

`func NewParamsPropertyInfraInventoryCriteriaInner() *ParamsPropertyInfraInventoryCriteriaInner`

NewParamsPropertyInfraInventoryCriteriaInner instantiates a new ParamsPropertyInfraInventoryCriteriaInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertyInfraInventoryCriteriaInnerWithDefaults

`func NewParamsPropertyInfraInventoryCriteriaInnerWithDefaults() *ParamsPropertyInfraInventoryCriteriaInner`

NewParamsPropertyInfraInventoryCriteriaInnerWithDefaults instantiates a new ParamsPropertyInfraInventoryCriteriaInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetric

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetMetric() string`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetMetricOk() (*string, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetMetric(v string)`

SetMetric sets Metric field to given value.

### HasMetric

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasMetric() bool`

HasMetric returns a boolean if a field has been set.

### GetTimeSize

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetTimeSize() float32`

GetTimeSize returns the TimeSize field if non-nil, zero value otherwise.

### GetTimeSizeOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetTimeSizeOk() (*float32, bool)`

GetTimeSizeOk returns a tuple with the TimeSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeSize

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetTimeSize(v float32)`

SetTimeSize sets TimeSize field to given value.

### HasTimeSize

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasTimeSize() bool`

HasTimeSize returns a boolean if a field has been set.

### GetTimeUnit

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetTimeUnit() string`

GetTimeUnit returns the TimeUnit field if non-nil, zero value otherwise.

### GetTimeUnitOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetTimeUnitOk() (*string, bool)`

GetTimeUnitOk returns a tuple with the TimeUnit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeUnit

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetTimeUnit(v string)`

SetTimeUnit sets TimeUnit field to given value.

### HasTimeUnit

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasTimeUnit() bool`

HasTimeUnit returns a boolean if a field has been set.

### GetSourceId

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetSourceId() string`

GetSourceId returns the SourceId field if non-nil, zero value otherwise.

### GetSourceIdOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetSourceIdOk() (*string, bool)`

GetSourceIdOk returns a tuple with the SourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceId

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetSourceId(v string)`

SetSourceId sets SourceId field to given value.

### HasSourceId

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasSourceId() bool`

HasSourceId returns a boolean if a field has been set.

### GetThreshold

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetThreshold() []float32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetThresholdOk() (*[]float32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetThreshold(v []float32)`

SetThreshold sets Threshold field to given value.

### HasThreshold

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasThreshold() bool`

HasThreshold returns a boolean if a field has been set.

### GetComparator

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetComparator() string`

GetComparator returns the Comparator field if non-nil, zero value otherwise.

### GetComparatorOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetComparatorOk() (*string, bool)`

GetComparatorOk returns a tuple with the Comparator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComparator

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetComparator(v string)`

SetComparator sets Comparator field to given value.

### HasComparator

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasComparator() bool`

HasComparator returns a boolean if a field has been set.

### GetCustomMetric

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetCustomMetric() ParamsPropertyInfraInventoryCriteriaInnerCustomMetric`

GetCustomMetric returns the CustomMetric field if non-nil, zero value otherwise.

### GetCustomMetricOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetCustomMetricOk() (*ParamsPropertyInfraInventoryCriteriaInnerCustomMetric, bool)`

GetCustomMetricOk returns a tuple with the CustomMetric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomMetric

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetCustomMetric(v ParamsPropertyInfraInventoryCriteriaInnerCustomMetric)`

SetCustomMetric sets CustomMetric field to given value.

### HasCustomMetric

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasCustomMetric() bool`

HasCustomMetric returns a boolean if a field has been set.

### GetWarningThreshold

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetWarningThreshold() []float32`

GetWarningThreshold returns the WarningThreshold field if non-nil, zero value otherwise.

### GetWarningThresholdOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetWarningThresholdOk() (*[]float32, bool)`

GetWarningThresholdOk returns a tuple with the WarningThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWarningThreshold

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetWarningThreshold(v []float32)`

SetWarningThreshold sets WarningThreshold field to given value.

### HasWarningThreshold

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasWarningThreshold() bool`

HasWarningThreshold returns a boolean if a field has been set.

### GetWarningComparator

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetWarningComparator() string`

GetWarningComparator returns the WarningComparator field if non-nil, zero value otherwise.

### GetWarningComparatorOk

`func (o *ParamsPropertyInfraInventoryCriteriaInner) GetWarningComparatorOk() (*string, bool)`

GetWarningComparatorOk returns a tuple with the WarningComparator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWarningComparator

`func (o *ParamsPropertyInfraInventoryCriteriaInner) SetWarningComparator(v string)`

SetWarningComparator sets WarningComparator field to given value.

### HasWarningComparator

`func (o *ParamsPropertyInfraInventoryCriteriaInner) HasWarningComparator() bool`

HasWarningComparator returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



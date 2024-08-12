# ParamsPropertySloBurnRate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SloId** | Pointer to **string** | The SLO identifier used by the rule | [optional] 
**BurnRateThreshold** | Pointer to **float32** | The burn rate threshold used to trigger the alert | [optional] 
**MaxBurnRateThreshold** | Pointer to **float32** | The maximum burn rate threshold value defined by the SLO error budget | [optional] 
**LongWindow** | Pointer to [**ParamsPropertySloBurnRateLongWindow**](ParamsPropertySloBurnRateLongWindow.md) |  | [optional] 
**ShortWindow** | Pointer to [**ParamsPropertySloBurnRateShortWindow**](ParamsPropertySloBurnRateShortWindow.md) |  | [optional] 

## Methods

### NewParamsPropertySloBurnRate

`func NewParamsPropertySloBurnRate() *ParamsPropertySloBurnRate`

NewParamsPropertySloBurnRate instantiates a new ParamsPropertySloBurnRate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertySloBurnRateWithDefaults

`func NewParamsPropertySloBurnRateWithDefaults() *ParamsPropertySloBurnRate`

NewParamsPropertySloBurnRateWithDefaults instantiates a new ParamsPropertySloBurnRate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSloId

`func (o *ParamsPropertySloBurnRate) GetSloId() string`

GetSloId returns the SloId field if non-nil, zero value otherwise.

### GetSloIdOk

`func (o *ParamsPropertySloBurnRate) GetSloIdOk() (*string, bool)`

GetSloIdOk returns a tuple with the SloId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSloId

`func (o *ParamsPropertySloBurnRate) SetSloId(v string)`

SetSloId sets SloId field to given value.

### HasSloId

`func (o *ParamsPropertySloBurnRate) HasSloId() bool`

HasSloId returns a boolean if a field has been set.

### GetBurnRateThreshold

`func (o *ParamsPropertySloBurnRate) GetBurnRateThreshold() float32`

GetBurnRateThreshold returns the BurnRateThreshold field if non-nil, zero value otherwise.

### GetBurnRateThresholdOk

`func (o *ParamsPropertySloBurnRate) GetBurnRateThresholdOk() (*float32, bool)`

GetBurnRateThresholdOk returns a tuple with the BurnRateThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBurnRateThreshold

`func (o *ParamsPropertySloBurnRate) SetBurnRateThreshold(v float32)`

SetBurnRateThreshold sets BurnRateThreshold field to given value.

### HasBurnRateThreshold

`func (o *ParamsPropertySloBurnRate) HasBurnRateThreshold() bool`

HasBurnRateThreshold returns a boolean if a field has been set.

### GetMaxBurnRateThreshold

`func (o *ParamsPropertySloBurnRate) GetMaxBurnRateThreshold() float32`

GetMaxBurnRateThreshold returns the MaxBurnRateThreshold field if non-nil, zero value otherwise.

### GetMaxBurnRateThresholdOk

`func (o *ParamsPropertySloBurnRate) GetMaxBurnRateThresholdOk() (*float32, bool)`

GetMaxBurnRateThresholdOk returns a tuple with the MaxBurnRateThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxBurnRateThreshold

`func (o *ParamsPropertySloBurnRate) SetMaxBurnRateThreshold(v float32)`

SetMaxBurnRateThreshold sets MaxBurnRateThreshold field to given value.

### HasMaxBurnRateThreshold

`func (o *ParamsPropertySloBurnRate) HasMaxBurnRateThreshold() bool`

HasMaxBurnRateThreshold returns a boolean if a field has been set.

### GetLongWindow

`func (o *ParamsPropertySloBurnRate) GetLongWindow() ParamsPropertySloBurnRateLongWindow`

GetLongWindow returns the LongWindow field if non-nil, zero value otherwise.

### GetLongWindowOk

`func (o *ParamsPropertySloBurnRate) GetLongWindowOk() (*ParamsPropertySloBurnRateLongWindow, bool)`

GetLongWindowOk returns a tuple with the LongWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLongWindow

`func (o *ParamsPropertySloBurnRate) SetLongWindow(v ParamsPropertySloBurnRateLongWindow)`

SetLongWindow sets LongWindow field to given value.

### HasLongWindow

`func (o *ParamsPropertySloBurnRate) HasLongWindow() bool`

HasLongWindow returns a boolean if a field has been set.

### GetShortWindow

`func (o *ParamsPropertySloBurnRate) GetShortWindow() ParamsPropertySloBurnRateShortWindow`

GetShortWindow returns the ShortWindow field if non-nil, zero value otherwise.

### GetShortWindowOk

`func (o *ParamsPropertySloBurnRate) GetShortWindowOk() (*ParamsPropertySloBurnRateShortWindow, bool)`

GetShortWindowOk returns a tuple with the ShortWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShortWindow

`func (o *ParamsPropertySloBurnRate) SetShortWindow(v ParamsPropertySloBurnRateShortWindow)`

SetShortWindow sets ShortWindow field to given value.

### HasShortWindow

`func (o *ParamsPropertySloBurnRate) HasShortWindow() bool`

HasShortWindow returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



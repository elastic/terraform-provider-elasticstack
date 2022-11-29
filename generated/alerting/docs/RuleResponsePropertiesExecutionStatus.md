# RuleResponsePropertiesExecutionStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastDuration** | Pointer to **int32** |  | [optional] 
**LastExecutionDate** | Pointer to **time.Time** |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 

## Methods

### NewRuleResponsePropertiesExecutionStatus

`func NewRuleResponsePropertiesExecutionStatus() *RuleResponsePropertiesExecutionStatus`

NewRuleResponsePropertiesExecutionStatus instantiates a new RuleResponsePropertiesExecutionStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRuleResponsePropertiesExecutionStatusWithDefaults

`func NewRuleResponsePropertiesExecutionStatusWithDefaults() *RuleResponsePropertiesExecutionStatus`

NewRuleResponsePropertiesExecutionStatusWithDefaults instantiates a new RuleResponsePropertiesExecutionStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLastDuration

`func (o *RuleResponsePropertiesExecutionStatus) GetLastDuration() int32`

GetLastDuration returns the LastDuration field if non-nil, zero value otherwise.

### GetLastDurationOk

`func (o *RuleResponsePropertiesExecutionStatus) GetLastDurationOk() (*int32, bool)`

GetLastDurationOk returns a tuple with the LastDuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastDuration

`func (o *RuleResponsePropertiesExecutionStatus) SetLastDuration(v int32)`

SetLastDuration sets LastDuration field to given value.

### HasLastDuration

`func (o *RuleResponsePropertiesExecutionStatus) HasLastDuration() bool`

HasLastDuration returns a boolean if a field has been set.

### GetLastExecutionDate

`func (o *RuleResponsePropertiesExecutionStatus) GetLastExecutionDate() time.Time`

GetLastExecutionDate returns the LastExecutionDate field if non-nil, zero value otherwise.

### GetLastExecutionDateOk

`func (o *RuleResponsePropertiesExecutionStatus) GetLastExecutionDateOk() (*time.Time, bool)`

GetLastExecutionDateOk returns a tuple with the LastExecutionDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastExecutionDate

`func (o *RuleResponsePropertiesExecutionStatus) SetLastExecutionDate(v time.Time)`

SetLastExecutionDate sets LastExecutionDate field to given value.

### HasLastExecutionDate

`func (o *RuleResponsePropertiesExecutionStatus) HasLastExecutionDate() bool`

HasLastExecutionDate returns a boolean if a field has been set.

### GetStatus

`func (o *RuleResponsePropertiesExecutionStatus) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *RuleResponsePropertiesExecutionStatus) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *RuleResponsePropertiesExecutionStatus) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *RuleResponsePropertiesExecutionStatus) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



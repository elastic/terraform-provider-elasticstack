# GetRuleTypes200ResponseInnerAlertsMappings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FieldMap** | Pointer to  | Mapping information for each field supported in alerts as data documents for this rule type. For more information about mapping parameters, refer to the Elasticsearch documentation.  | [optional] 

## Methods

### NewGetRuleTypes200ResponseInnerAlertsMappings

`func NewGetRuleTypes200ResponseInnerAlertsMappings() *GetRuleTypes200ResponseInnerAlertsMappings`

NewGetRuleTypes200ResponseInnerAlertsMappings instantiates a new GetRuleTypes200ResponseInnerAlertsMappings object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetRuleTypes200ResponseInnerAlertsMappingsWithDefaults

`func NewGetRuleTypes200ResponseInnerAlertsMappingsWithDefaults() *GetRuleTypes200ResponseInnerAlertsMappings`

NewGetRuleTypes200ResponseInnerAlertsMappingsWithDefaults instantiates a new GetRuleTypes200ResponseInnerAlertsMappings object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFieldMap

`func (o *GetRuleTypes200ResponseInnerAlertsMappings) GetFieldMap() map[string]FieldmapProperties`

GetFieldMap returns the FieldMap field if non-nil, zero value otherwise.

### GetFieldMapOk

`func (o *GetRuleTypes200ResponseInnerAlertsMappings) GetFieldMapOk() (*map[string]FieldmapProperties, bool)`

GetFieldMapOk returns a tuple with the FieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldMap

`func (o *GetRuleTypes200ResponseInnerAlertsMappings) SetFieldMap(v map[string]FieldmapProperties)`

SetFieldMap sets FieldMap field to given value.

### HasFieldMap

`func (o *GetRuleTypes200ResponseInnerAlertsMappings) HasFieldMap() bool`

HasFieldMap returns a boolean if a field has been set.

### SetFieldMapNil

`func (o *GetRuleTypes200ResponseInnerAlertsMappings) SetFieldMapNil(b bool)`

 SetFieldMapNil sets the value for FieldMap to be an explicit nil

### UnsetFieldMap
`func (o *GetRuleTypes200ResponseInnerAlertsMappings) UnsetFieldMap()`

UnsetFieldMap ensures that no value is present for FieldMap, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



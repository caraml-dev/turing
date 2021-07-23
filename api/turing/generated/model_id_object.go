/*
 * Turing Minimal Openapi Spec for SDK
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.1
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// IdObject struct for IdObject
type IdObject struct {
	Id *int32 `json:"id,omitempty"`
}

// NewIdObject instantiates a new IdObject object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIdObject() *IdObject {
	this := IdObject{}
	return &this
}

// NewIdObjectWithDefaults instantiates a new IdObject object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIdObjectWithDefaults() *IdObject {
	this := IdObject{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *IdObject) GetId() int32 {
	if o == nil || o.Id == nil {
		var ret int32
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IdObject) GetIdOk() (*int32, bool) {
	if o == nil || o.Id == nil {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *IdObject) HasId() bool {
	if o != nil && o.Id != nil {
		return true
	}

	return false
}

// SetId gets a reference to the given int32 and assigns it to the Id field.
func (o *IdObject) SetId(v int32) {
	o.Id = &v
}

func (o IdObject) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Id != nil {
		toSerialize["id"] = o.Id
	}
	return json.Marshal(toSerialize)
}

type NullableIdObject struct {
	value *IdObject
	isSet bool
}

func (v NullableIdObject) Get() *IdObject {
	return v.value
}

func (v *NullableIdObject) Set(val *IdObject) {
	v.value = val
	v.isSet = true
}

func (v NullableIdObject) IsSet() bool {
	return v.isSet
}

func (v *NullableIdObject) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIdObject(val *IdObject) *NullableIdObject {
	return &NullableIdObject{value: val, isSet: true}
}

func (v NullableIdObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIdObject) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

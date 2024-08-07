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

// EnsemblingJobEnsemblerSpec struct for EnsemblingJobEnsemblerSpec
type EnsemblingJobEnsemblerSpec struct {
	Uri string `json:"uri"`
	Result EnsemblingJobEnsemblerSpecResult `json:"result"`
}

// NewEnsemblingJobEnsemblerSpec instantiates a new EnsemblingJobEnsemblerSpec object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEnsemblingJobEnsemblerSpec(uri string, result EnsemblingJobEnsemblerSpecResult) *EnsemblingJobEnsemblerSpec {
	this := EnsemblingJobEnsemblerSpec{}
	this.Uri = uri
	this.Result = result
	return &this
}

// NewEnsemblingJobEnsemblerSpecWithDefaults instantiates a new EnsemblingJobEnsemblerSpec object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEnsemblingJobEnsemblerSpecWithDefaults() *EnsemblingJobEnsemblerSpec {
	this := EnsemblingJobEnsemblerSpec{}
	var uri string = ""
	this.Uri = uri
	return &this
}

// GetUri returns the Uri field value
func (o *EnsemblingJobEnsemblerSpec) GetUri() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Uri
}

// GetUriOk returns a tuple with the Uri field value
// and a boolean to check if the value has been set.
func (o *EnsemblingJobEnsemblerSpec) GetUriOk() (*string, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Uri, true
}

// SetUri sets field value
func (o *EnsemblingJobEnsemblerSpec) SetUri(v string) {
	o.Uri = v
}

// GetResult returns the Result field value
func (o *EnsemblingJobEnsemblerSpec) GetResult() EnsemblingJobEnsemblerSpecResult {
	if o == nil {
		var ret EnsemblingJobEnsemblerSpecResult
		return ret
	}

	return o.Result
}

// GetResultOk returns a tuple with the Result field value
// and a boolean to check if the value has been set.
func (o *EnsemblingJobEnsemblerSpec) GetResultOk() (*EnsemblingJobEnsemblerSpecResult, bool) {
	if o == nil  {
		return nil, false
	}
	return &o.Result, true
}

// SetResult sets field value
func (o *EnsemblingJobEnsemblerSpec) SetResult(v EnsemblingJobEnsemblerSpecResult) {
	o.Result = v
}

func (o EnsemblingJobEnsemblerSpec) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["uri"] = o.Uri
	}
	if true {
		toSerialize["result"] = o.Result
	}
	return json.Marshal(toSerialize)
}

type NullableEnsemblingJobEnsemblerSpec struct {
	value *EnsemblingJobEnsemblerSpec
	isSet bool
}

func (v NullableEnsemblingJobEnsemblerSpec) Get() *EnsemblingJobEnsemblerSpec {
	return v.value
}

func (v *NullableEnsemblingJobEnsemblerSpec) Set(val *EnsemblingJobEnsemblerSpec) {
	v.value = val
	v.isSet = true
}

func (v NullableEnsemblingJobEnsemblerSpec) IsSet() bool {
	return v.isSet
}

func (v *NullableEnsemblingJobEnsemblerSpec) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnsemblingJobEnsemblerSpec(val *EnsemblingJobEnsemblerSpec) *NullableEnsemblingJobEnsemblerSpec {
	return &NullableEnsemblingJobEnsemblerSpec{value: val, isSet: true}
}

func (v NullableEnsemblingJobEnsemblerSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnsemblingJobEnsemblerSpec) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}



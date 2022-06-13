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
	"fmt"
)

// EnsemblerConfigKind the model 'EnsemblerConfigKind'
type EnsemblerConfigKind string

// List of EnsemblerConfigKind
const (
	ENSEMBLERCONFIGKIND_BATCH_ENSEMBLING_JOB EnsemblerConfigKind = "BatchEnsemblingJob"
)

var allowedEnsemblerConfigKindEnumValues = []EnsemblerConfigKind{
	"BatchEnsemblingJob",
}

func (v *EnsemblerConfigKind) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := EnsemblerConfigKind(value)
	for _, existing := range allowedEnsemblerConfigKindEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid EnsemblerConfigKind", value)
}

// NewEnsemblerConfigKindFromValue returns a pointer to a valid EnsemblerConfigKind
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewEnsemblerConfigKindFromValue(v string) (*EnsemblerConfigKind, error) {
	ev := EnsemblerConfigKind(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for EnsemblerConfigKind: valid values are %v", v, allowedEnsemblerConfigKindEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v EnsemblerConfigKind) IsValid() bool {
	for _, existing := range allowedEnsemblerConfigKindEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to EnsemblerConfigKind value
func (v EnsemblerConfigKind) Ptr() *EnsemblerConfigKind {
	return &v
}

type NullableEnsemblerConfigKind struct {
	value *EnsemblerConfigKind
	isSet bool
}

func (v NullableEnsemblerConfigKind) Get() *EnsemblerConfigKind {
	return v.value
}

func (v *NullableEnsemblerConfigKind) Set(val *EnsemblerConfigKind) {
	v.value = val
	v.isSet = true
}

func (v NullableEnsemblerConfigKind) IsSet() bool {
	return v.isSet
}

func (v *NullableEnsemblerConfigKind) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnsemblerConfigKind(val *EnsemblerConfigKind) *NullableEnsemblerConfigKind {
	return &NullableEnsemblerConfigKind{value: val, isSet: true}
}

func (v NullableEnsemblerConfigKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnsemblerConfigKind) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

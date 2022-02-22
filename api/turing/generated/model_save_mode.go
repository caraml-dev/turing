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

// SaveMode the model 'SaveMode'
type SaveMode string

// List of SaveMode
const (
	SAVEMODE_ERRORIFEXISTS SaveMode = "ERRORIFEXISTS"
	SAVEMODE_OVERWRITE     SaveMode = "OVERWRITE"
	SAVEMODE_APPEND        SaveMode = "APPEND"
	SAVEMODE_IGNORE        SaveMode = "IGNORE"
)

var allowedSaveModeEnumValues = []SaveMode{
	"ERRORIFEXISTS",
	"OVERWRITE",
	"APPEND",
	"IGNORE",
}

func (v *SaveMode) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := SaveMode(value)
	for _, existing := range allowedSaveModeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid SaveMode", value)
}

// NewSaveModeFromValue returns a pointer to a valid SaveMode
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewSaveModeFromValue(v string) (*SaveMode, error) {
	ev := SaveMode(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for SaveMode: valid values are %v", v, allowedSaveModeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v SaveMode) IsValid() bool {
	for _, existing := range allowedSaveModeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to SaveMode value
func (v SaveMode) Ptr() *SaveMode {
	return &v
}

type NullableSaveMode struct {
	value *SaveMode
	isSet bool
}

func (v NullableSaveMode) Get() *SaveMode {
	return v.value
}

func (v *NullableSaveMode) Set(val *SaveMode) {
	v.value = val
	v.isSet = true
}

func (v NullableSaveMode) IsSet() bool {
	return v.isSet
}

func (v *NullableSaveMode) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSaveMode(val *SaveMode) *NullableSaveMode {
	return &NullableSaveMode{value: val, isSet: true}
}

func (v NullableSaveMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSaveMode) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

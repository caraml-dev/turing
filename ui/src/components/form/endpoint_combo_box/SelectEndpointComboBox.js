import React from "react";
import { isValidUrl } from "../../../utils/validation";
import { EuiComboBoxSelect } from "../combo_box/EuiComboBoxSelect";

export const SelectEndpointComboBox = ({
  value,
  onChange,
  options,
  ...props
}) => {
  const onCreateOption = value => {
    if (!isValidUrl(value)) {
      return false;
    }
    onChange(value);
  };

  return (
    <EuiComboBoxSelect
      fullWidth={props.fullWidth}
      placeholder={props.placeholder}
      value={value}
      options={options}
      onChange={onChange}
      onCreateOption={onCreateOption}
    />
  );
};

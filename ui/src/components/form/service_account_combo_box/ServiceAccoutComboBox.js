import React, { useMemo } from "react";
import { EuiComboBoxSelect } from "../combo_box/EuiComboBoxSelect";
import serviceAccIcon from "../../../assets/icons/service-account.svg";

export const ServiceAccountComboBox = ({
  value,
  secrets,
  onChange,
  ...props
}) => {
  const secretsOptions = useMemo(
    () =>
      secrets
        ? secrets.map(s => ({
            label: s.name,
            icon: serviceAccIcon
          }))
        : [],
    [secrets]
  );

  return (
    <EuiComboBoxSelect
      className="euiComboBox--serviceAccount"
      fullWidth
      placeholder="Select Service Account..."
      isInvalid={props.isInvalid}
      options={secretsOptions}
      value={value}
      onChange={value => onChange(value || "")}
    />
  );
};

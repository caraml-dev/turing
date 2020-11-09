import React, { useMemo } from "react";
import { EuiFieldNumber, EuiSuperSelect } from "@elastic/eui";
import { durationOptions } from "../../../../config";
import { intValue } from "../metric_panel/MetricPanel";

export const FieldDuration = ({ value, onChange }) => {
  const { duration, unit } = useMemo(
    () => ({
      duration: !!value ? value.substr(0, value.length - 1) : 0,
      unit: !!value ? value.substr(-1, 1) : "m"
    }),
    [value]
  );

  const selectedOption = durationOptions.find(option => option.value === unit);

  return (
    <EuiFieldNumber
      min={0}
      value={duration || 0}
      onChange={e => onChange(`${intValue(e) || 0}${unit}`)}
      append={
        <EuiSuperSelect
          itemLayoutAlign="top"
          hasDividers
          options={durationOptions}
          valueOfSelected={selectedOption ? selectedOption.value : ""}
          onChange={value => onChange(`${duration}${value}`)}
        />
      }
    />
  );
};

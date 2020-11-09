import React from "react";
import { EuiFieldNumber, EuiFormLabel } from "@elastic/eui";

const durationRegex = /([0-9]+)(ms|s|m|h)$/;

const parseDuration = value => {
  if (value) {
    let matches = `${value}`.match(durationRegex);
    if (matches) {
      return [parseInt(matches[1]), matches[2]];
    }
  }
  return undefined;
};

export const EuiFieldDuration = props => {
  const [value, timeUnit] = parseDuration(props.value) || [0, "ms"];

  return (
    <EuiFieldNumber
      {...props}
      min={0}
      value={value}
      onChange={e => props.onChange(`${e.target.value}${timeUnit}`)}
      append={<EuiFormLabel>{timeUnit}</EuiFormLabel>}
    />
  );
};

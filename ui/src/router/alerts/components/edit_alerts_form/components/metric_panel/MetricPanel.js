import React, { Fragment } from "react";
import {
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSpacer,
  EuiSwitch,
  EuiText
} from "@elastic/eui";
import { Panel } from "../../../../../components/form/components/Panel";
import "./MetricPanel.scss";
import { FieldDuration } from "../field_duration/FieldDuration";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";

export const intValue = e => {
  if (e && e.target.value) {
    const intVal = parseInt(e.target.value);
    if (!isNaN(intVal)) {
      return intVal;
    }
  }
  return undefined;
};

export const MetricPanel = ({
  title,
  comparator,
  unit,
  alert = {},
  onChangeHandler,
  errors = {},
  isExpanded,
  toggleExpanded
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <Panel contentWidth="65%">
      <EuiSpacer size="m" />
      <EuiFlexGroup alignItems="center" justifyContent="spaceBetween">
        <EuiFlexItem>
          <EuiText size="m">
            <h4>{title}</h4>
          </EuiText>
          <EuiSpacer size="s" />
          <EuiText size="s" color="subdued">
            {isExpanded ? (
              <Fragment>
                The team will be notified if {title} is <b>{comparator}</b> than
                threshold.
              </Fragment>
            ) : (
              <Fragment>Alert for {title} is disabled.</Fragment>
            )}
          </EuiText>
        </EuiFlexItem>
        <EuiFlexItem grow={false}>
          <EuiSpacer size="s" />
          <EuiSwitch label="" checked={isExpanded} onChange={toggleExpanded} />
        </EuiFlexItem>
      </EuiFlexGroup>
      {isExpanded && (
        <Fragment>
          <EuiSpacer size="l" />
          <EuiSpacer size="l" />
          <EuiFlexGroup direction="column">
            <EuiFlexItem>
              <EuiFlexGroup>
                <EuiFlexItem>
                  <EuiFormRow
                    label={"Warning threshold"}
                    isInvalid={!!errors.warning_threshold || !!errors.overall}
                    error={errors.warning_threshold}>
                    <EuiFieldNumber
                      min={0}
                      value={alert.warning_threshold || 0}
                      onChange={e => onChange("warning_threshold")(intValue(e))}
                      append={unit}
                    />
                  </EuiFormRow>
                </EuiFlexItem>
                <EuiFlexItem>
                  <EuiFormRow
                    label={"Critical threshold"}
                    isInvalid={!!errors.critical_threshold || !!errors.overall}
                    error={errors.critical_threshold}>
                    <EuiFieldNumber
                      min={0}
                      value={alert.critical_threshold || 0}
                      onChange={e =>
                        onChange("critical_threshold")(intValue(e))
                      }
                      append={unit}
                    />
                  </EuiFormRow>
                </EuiFlexItem>
              </EuiFlexGroup>
            </EuiFlexItem>

            <EuiFlexItem>
              <EuiFormRow
                className="durationRow"
                label={"Duration"}
                isInvalid={!!errors.duration}
                error={errors.duration}>
                <FieldDuration
                  value={alert.duration}
                  onChange={onChange("duration")}
                />
              </EuiFormRow>
            </EuiFlexItem>

            {errors.overall && (
              <EuiFlexItem>
                <EuiText size="s" color="danger">
                  {errors.overall}
                </EuiText>
              </EuiFlexItem>
            )}
          </EuiFlexGroup>
        </Fragment>
      )}
      <EuiSpacer size="s" />
    </Panel>
  );
};

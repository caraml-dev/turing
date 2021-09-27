import React, { useEffect, useState } from "react";
import {
  EuiButtonIcon,
  EuiCard,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
} from "@elastic/eui";
import { SelectEndpointComboBox } from "../../../../../../components/form/endpoint_combo_box/SelectEndpointComboBox";
import { EuiFieldDuration } from "../../../../../../components/form/field_duration/EuiFieldDuration";
import { get } from "../../../../../../components/form/utils";
import "./RouteCard.scss";

export const RouteCard = ({
  route,
  isDefault,
  endpointOptions,
  onChange,
  onSelect,
  onDelete,
  errors,
}) => {
  const [endpointsMap, setEndpointsMap] = useState({});

  useEffect(() => {
    setEndpointsMap(
      endpointOptions
        .flatMap((option) => option.options || [option])
        .reduce((dict, option) => {
          dict[option.label] = option;
          return dict;
        }, {})
    );
  }, [endpointOptions, setEndpointsMap]);

  return (
    <EuiCard
      className="euiCard--routeCard"
      title=""
      description=""
      textAlign="left"
      selectable={{
        children: isDefault ? "Default" : "Set Default",
        onClick: onSelect,
        isSelected: isDefault,
      }}
    >
      <EuiForm onClick={(e) => e.stopPropagation()}>
        <EuiFlexGroup
          className="euiFlexGroup--removeButton"
          justifyContent="flexEnd"
          gutterSize="none"
          direction="row"
        >
          <EuiFlexItem grow={false}>
            {!isDefault ? (
              <EuiButtonIcon
                iconType="cross"
                onClick={onDelete}
                aria-label="delete-route"
              />
            ) : (
              <EuiIcon type="empty" size="l" />
            )}
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiFormRow
          label="Endpoint *"
          isInvalid={!!get(errors, "endpoint")}
          error={get(errors, "endpoint")}
          fullWidth
        >
          <SelectEndpointComboBox
            fullWidth
            placeholder="http://models.internal/predict"
            value={route.endpoint}
            options={endpointOptions}
            onChange={(endpoint) => {
              onChange({
                ...route,
                endpoint,
                annotations: (endpointsMap[endpoint] || {}).annotations,
              });
            }}
            isInvalid={!!get(errors, "endpoint")}
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label="Route Id *"
              isInvalid={!!get(errors, "id")}
              error={get(errors, "id")}
            >
              <EuiFieldText
                placeholder="control"
                value={route.id}
                onChange={(e) => onChange({ ...route, id: e.target.value })}
                isInvalid={!!get(errors, "id")}
                aria-label="route-id"
              />
            </EuiFormRow>
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiFormRow
              label="Timeout *"
              isInvalid={!!get(errors, "timeout")}
              error={get(errors, "timeout")}
            >
              <EuiFieldDuration
                value={route.timeout}
                max={10000}
                onChange={(value) => onChange({ ...route, timeout: value })}
                isInvalid={!!get(errors, "timeout")}
                aria-label="route-timeout"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </EuiCard>
  );
};

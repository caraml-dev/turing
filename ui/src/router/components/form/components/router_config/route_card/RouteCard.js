import React, { useEffect, useState } from "react";
import {
  EuiButtonIcon,
  EuiCard,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
} from "@elastic/eui";
import { SelectEndpointComboBox } from "../../../../../../components/form/endpoint_combo_box/SelectEndpointComboBox";
import { EuiFieldDuration } from "../../../../../../components/form/field_duration/EuiFieldDuration";
import { get } from "../../../../../../components/form/utils";
import "./RouteCard.scss";
import { EuiComboBoxSelect } from "../../../../../../components/form/combo_box/EuiComboBoxSelect";

export const RouteCard = ({
  route,
  protocol,
  endpointOptions,
  onChange,
  onDelete,
  errors,
}) => {
  const [endpointsMap, setEndpointsMap] = useState({});

  const upiServiceMethodOptions = [{
    label: "/caraml.upi.v1.UniversalPredictionService/PredictValues",
    value: "/caraml.upi.v1.UniversalPredictionService/PredictValues"
  }]

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
      textAlign="left">
      <EuiForm onClick={(e) => e.stopPropagation()}>
        <EuiFlexGroup
          className="euiFlexGroup--removeButton"
          justifyContent="flexEnd"
          gutterSize="none"
          direction="row">
          {!!onDelete && (
            <EuiFlexItem grow={false}>
              <EuiButtonIcon
                iconType="cross"
                onClick={onDelete}
                aria-label="delete-route"
              />
            </EuiFlexItem>
          )}
        </EuiFlexGroup>

        <EuiFormRow
          label="Endpoint *"
          isInvalid={!!get(errors, "endpoint")}
          error={get(errors, "endpoint")}
          fullWidth>
          <SelectEndpointComboBox
            fullWidth
            placeholder={protocol === "HTTP_JSON" ? "http://models.internal/predict": "models.internal:80"} 
            protocol={protocol}
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

        {
          protocol === "UPI_V1" && 
          (
            <EuiFormRow
              label="Service Method *"
              isInvalid={!!get(errors, "service_method")}
              error={get(errors, "service_method")}
              fullWidth>
              <EuiComboBoxSelect
                fullWidth
                placeholder="/caraml.upi.v1.UniversalPredictionService/PredictValues"
                value={route.service_method}
                options={upiServiceMethodOptions}
                onChange={(service_method) => {
                  onChange({
                    ...route,
                    service_method,
                  });
                }}
                onCreateOption={(service_method) => {
                  onChange({
                    ...route,
                    service_method,
                  });
                }}
              />
          </EuiFormRow>
          )
        }

        <EuiSpacer size="m" />

        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label="Route Id *"
              isInvalid={!!get(errors, "id")}
              error={get(errors, "id")}>
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
              error={get(errors, "timeout")}>
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

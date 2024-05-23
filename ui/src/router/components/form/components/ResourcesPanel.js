import React, { useMemo } from "react";
import { Panel } from "./Panel";
import {
  EuiDualRange,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiAccordion,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import { CPULimitsFormGroup } from "./CPULimitsFormGroup";

export const ResourcesPanel = ({
  resourcesConfig,
  onChangeHandler,
  errors = {},
  maxAllowedReplica,
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const replicasError = useMemo(
    () => [...(errors.min_replica || []), ...(errors.max_replica || [])],
    [errors.min_replica, errors.max_replica]
  );

  return (
    <Panel title="Resources">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="CPU*"
                  content="Specify the total amount of CPU available for the component"
                />
              }
              isInvalid={!!errors.cpu_request}
              error={errors.cpu_request}
              fullWidth>
              <EuiFieldText
                placeholder="500m"
                value={resourcesConfig.cpu_request}
                onChange={(e) => onChange("cpu_request")(e.target.value)}
                isInvalid={!!errors.cpu_request}
                name="cpu"
                fullWidth
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Memory*"
                  content="Specify the total amount of RAM available for the component"
                />
              }
              isInvalid={!!errors.memory_request}
              error={errors.memory_request}
              fullWidth>
              <EuiFieldText
                placeholder="500Mi"
                value={resourcesConfig.memory_request}
                onChange={(e) => onChange("memory_request")(e.target.value)}
                isInvalid={!!errors.memory_request}
                name="memory"
                fullWidth
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiFormRow
          label={
            <FormLabelWithToolTip
              label="Min/Max Replicas*"
              content="Specify the min/max number of replicas for your deployment. We will take care about auto-scaling for you"
            />
          }
          isInvalid={!!replicasError.length}
          error={replicasError}
          fullWidth>
          <EuiDualRange
            fullWidth
            min={0}
            max={maxAllowedReplica}
            showInput
            showLabels
            value={[
              resourcesConfig.min_replica || 0,
              resourcesConfig.max_replica || 0,
            ]}
            onChange={([min_replica, max_replica]) => {
              onChange("min_replica")(parseInt(min_replica));
              onChange("max_replica")(parseInt(max_replica));
            }}
            isInvalid={!!replicasError.length}
            aria-label="autoscaling"
          />
        </EuiFormRow>
        <EuiSpacer size="s" />
        <EuiAccordion
          id="adv config"
          buttonContent="Advanced configurations">
          <EuiSpacer size="s" />
          <CPULimitsFormGroup
            resourcesConfig={resourcesConfig}
            onChange={(e) => onChange("cpu_limit")(e.target.value)}
            errors={errors.cpu_limit}
          />
        </EuiAccordion>
      </EuiForm>
    </Panel>
  );
};

import React, {useState} from "react";
import {
  EuiFlexItem,
  EuiSwitch,
  EuiFlexGroup,
  EuiForm,
  EuiSpacer
} from "@elastic/eui";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import { get } from "../../../../components/form/utils";
import { AutoscalingPolicyPanel } from "./autoscaling_policy/AutoscalingPolicyPanel";
import { ResourcesPanel } from "./ResourcesPanel";
import {Panel} from "./Panel";
import {DefaultAutoscalingPolicyPanel} from "./DefaultAutoscalingPolicyPanel";


export const ResourcesAndAutoscalingPolicyPanel = ({
  data,
  onChangeHandler,
  errors = {},
  maxAllowedReplica,
}) => {
  const [useDefaultAutoscalingPolicy, setUseDefaultAutoscalingPolicy] = useState(true);

  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <>
      <EuiFlexItem grow={false}>
        <Panel title="Resource Usage">
          <EuiForm>
            <EuiFlexGroup direction="row">
              <EuiFlexItem>
                <EuiSpacer size={"m"}/>
                <EuiSwitch
                  label={useDefaultAutoscalingPolicy ? "Use default autoscaling policy" : "Customise resources manually"}
                  checked={useDefaultAutoscalingPolicy}
                  onChange={() => setUseDefaultAutoscalingPolicy(!useDefaultAutoscalingPolicy)}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </EuiForm>
        </Panel>
      </EuiFlexItem>
      {
        !useDefaultAutoscalingPolicy ?
          <>
            <EuiFlexItem grow={false}>
              <ResourcesPanel
                resourcesConfig={get(data, "config.resource_request")}
                onChangeHandler={onChange("resource_request")}
                maxAllowedReplica={maxAllowedReplica}
                errors={get(errors, "config.resource_request")}
              />
            </EuiFlexItem>
            <EuiFlexItem grow={false}>
              <AutoscalingPolicyPanel
                autoscalingPolicyConfig={get(data, "config.autoscaling_policy")}
                onChangeHandler={onChange("autoscaling_policy")}
                errors={get(errors, "config.autoscaling_policy")}
              />
            </EuiFlexItem>
          </> :
          <EuiFlexItem grow={false}>
            <DefaultAutoscalingPolicyPanel
              autoscalingPolicyConfig={get(data, "config.autoscaling_policy")}
              onChangeHandler={onChange("autoscaling_policy")}
              errors={get(errors, "config.autoscaling_policy")}
            />
          </EuiFlexItem>
      }
    </>
  );
};

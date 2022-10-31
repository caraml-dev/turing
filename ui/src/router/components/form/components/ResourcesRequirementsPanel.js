import React from "react";
import {
  EuiFlexItem,
} from "@elastic/eui";
import { AutoscalingPolicyPanel } from "./autoscaling_policy/AutoscalingPolicyPanel";
import { ResourcesPanel } from "./ResourcesPanel";
import { DefaultAutoscalingPolicyPanel } from "./DefaultAutoscalingPolicyPanel";
import { useToggle } from "@gojek/mlp-ui";

export const ResourcesRequirementsPanel = ({
  resources,
  resourcesOnChangeHandler,
  resourcesErrors,
  autoscalingPolicy,
  autoscalingPolicyOnChangeHandler,
  autoscalingPolicyErrors,
  maxAllowedReplica,
}) => {
  // Determine default state of switch depending on whether the default autoscaling policy has been set
  const [useDefaultAutoscalingPolicy, toggleUseDefaultAutoscalingPolicy] = useToggle(autoscalingPolicy.payload_size !== "");

  const onToggle = () => {
    if (useDefaultAutoscalingPolicy) {
      autoscalingPolicyOnChangeHandler({metric: "concurrency", target: "1", payload_size: null});
    } else {
      autoscalingPolicyOnChangeHandler({metric: null, target: null, payload_size: "200Mi"});
    }
    toggleUseDefaultAutoscalingPolicy();
  }

  return (
    <>
      <EuiFlexItem grow={false}>
        <DefaultAutoscalingPolicyPanel
          useDefaultAutoscalingPolicy={useDefaultAutoscalingPolicy}
          toggleUseDefaultAutoscalingPolicy={onToggle}
          autoscalingPolicyConfig={autoscalingPolicy}
          onChangeHandler={autoscalingPolicyOnChangeHandler}
          errors={autoscalingPolicyErrors}
        />
      </EuiFlexItem>
      {
        !useDefaultAutoscalingPolicy &&
          <>
            <EuiFlexItem grow={false}>
              <ResourcesPanel
                resourcesConfig={resources}
                onChangeHandler={resourcesOnChangeHandler}
                maxAllowedReplica={maxAllowedReplica}
                errors={resourcesErrors}
              />
            </EuiFlexItem>
            <EuiFlexItem grow={false}>
              <AutoscalingPolicyPanel
                autoscalingPolicyConfig={autoscalingPolicy}
                onChangeHandler={autoscalingPolicyOnChangeHandler}
                errors={autoscalingPolicyErrors}
              />
            </EuiFlexItem>
          </>
      }
    </>
  );
};

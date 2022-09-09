import React from "react";
import { EuiDescriptionList } from "@elastic/eui";

import { autoscalingPolicyOptions } from "../../form/components/autoscaling_policy/typeOptions";
import { ConfigMultiSectionPanel } from "../../../../components/config_multi_section_panel/ConfigMultiSectionPanel"

const ResourcesSection = ({
  resourceRequest: { cpu_request, memory_request, min_replica, max_replica },
}) => {
  const items = [
    {
      title: "CPU Request",
      description: cpu_request,
    },
    {
      title: "Memory Request",
      description: memory_request,
    },
    {
      title: "Min Replicas",
      description: min_replica,
    },
    {
      title: "Max Replicas",
      description: max_replica,
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
    />
  );
}

const AutoscalingPolicySection = ({
  autoscalingPolicy: { metric, target },
}) => {
  const selectedMetric = autoscalingPolicyOptions.find((e) => e.value === metric);
  const items = [
    {
      title: "Metric",
      description: selectedMetric?.inputDisplay,
    },
    {
      title: "Target",
      description: target,
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
    />
  );
};


export const ResourcesConfigTable = ({ componentName, autoscalingPolicy, resourceRequest }) => {
  const items = [
    {
      title: `${componentName} Resources`,
      children: <ResourcesSection resourceRequest={resourceRequest} />,
    },
    {
      title: "Autoscaling Policy",
      children: <AutoscalingPolicySection autoscalingPolicy={autoscalingPolicy} />,
    },
  ];

  return <ConfigMultiSectionPanel items={items} />
};

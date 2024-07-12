import {
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHorizontalRule,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import React from "react";

export const ResourceRequestConfigTable = ({
  resources: {
    driver_cpu_request,
    driver_memory_request,
    executor_cpu_request,
    executor_memory_request,
    executor_replica,
  },
}) => {
  const driver_items = [
    {
      title: "CPU Request",
      description: driver_cpu_request,
    },
    {
      title: "Memory Request",
      description: driver_memory_request,
    },
  ];

  const executor_items = [
    {
      title: "CPU Request",
      description: executor_cpu_request,
    },
    {
      title: "Memory Request",
      description: executor_memory_request,
    },
    {
      title: "Number of Replicas",
      description: executor_replica,
    },
  ];

  return (
    <EuiFlexGroup direction="column" gutterSize="none">
      <EuiFlexItem>
        <EuiTitle size="xxs">
          <span>
            <EuiTextColor color="success">Driver</EuiTextColor>
          </span>
        </EuiTitle>
        <EuiDescriptionList
          compressed
          textStyle="reverse"
          type="responsiveColumn"
          listItems={driver_items}
          columnWidths={[1, 7/3]}
        />
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiHorizontalRule size="full" margin="s" />
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiTitle size="xxs">
          <span>
            <EuiTextColor color="success">Executors</EuiTextColor>
          </span>
        </EuiTitle>
        <EuiDescriptionList
          compressed
          textStyle="reverse"
          type="responsiveColumn"
          listItems={executor_items}
          columnWidths={[1, 7/3]}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

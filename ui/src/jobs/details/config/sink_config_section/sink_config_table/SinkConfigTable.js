import React from "react";
import { EuiDescriptionList, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { SinkType } from "../../../../../services/job/SinkType";

export const SinkConfigTable = ({ sink, ensembler }) => {
  const items = [
    {
      title: "Type",
      description: SinkType.fromValue(sink.type).label,
    },
    {
      title: "Result Column Name",
      description: ensembler.result.column_name,
    },
    {
      title: "Result Type",
      description:
        ensembler.result.type === "ARRAY"
          ? `ARRAY(${ensembler.result.item_type})`
          : ensembler.result.type,
    },
    {
      title: "Persisted Columns",
      description: (
        <EuiFlexGroup direction="column" gutterSize="xxs">
          {sink.columns.map((item, idx) => (
            <EuiFlexItem key={idx}>
              {item}
              {idx + 1 < sink.columns.length && ","}
            </EuiFlexItem>
          ))}
        </EuiFlexGroup>
      ),
    },
    {
      title: "Save Mode",
      description: sink.save_mode,
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  );
};

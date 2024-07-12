import React from "react";
import { EuiDescriptionList } from "@elastic/eui";
import { SinkType } from "../../../../../services/job/SinkType";
import { CollapsibleBadgeGroup } from "../../../../../components/collapsible_badge_group/CollapsibleBadgeGroup";

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
        <CollapsibleBadgeGroup
          className="euiBadgeGroup---insideDescriptionList"
          items={sink.columns}
          minItemsCount={5}
        />
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
      columnWidths={[1, 7/3]}
    />
  );
};

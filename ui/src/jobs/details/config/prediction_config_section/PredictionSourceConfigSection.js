import { EuiFlexGroup, EuiFlexItem, EuiListGroup } from "@elastic/eui";
import { SourceType } from "../../../../services/job/SourceType";
import { BigQueryDatasetSectionPanel } from "../source_config_section/bq_config_section/BigQueryDatasetSection";
import React from "react";
import { ConfigMultiSectionPanel } from "../../../../components/config_multi_section_panel/ConfigMultiSectionPanel";

export const PredictionSourceConfigSection = ({ source }) => {
  const items = [
    {
      title: "Join Columns",
      children: (
        <EuiListGroup
          flush
          listItems={source.join_on.map((item) => ({ label: item }))}
        />
      ),
    },
    {
      title: "Prediction Features",
      children: (
        <EuiListGroup
          flush
          listItems={source.columns.map((item) => ({ label: item }))}
        />
      ),
    },
  ];

  return (
    <EuiFlexGroup direction="row">
      {
        /* eslint-disable-next-line eqeqeq */
        source.dataset.type == SourceType.BQ && (
          <EuiFlexItem grow={3}>
            <BigQueryDatasetSectionPanel bqConfig={source.dataset.bq_config} />
          </EuiFlexItem>
        )
      }

      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigMultiSectionPanel items={items} />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

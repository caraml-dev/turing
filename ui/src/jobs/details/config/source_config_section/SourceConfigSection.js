import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiListGroup } from "@elastic/eui";
import { SourceType } from "../../../../services/job/SourceType";
import { BigQueryDatasetSectionPanel } from "./bq_config_section/BigQueryDatasetSection";
import { ConfigSectionPanel } from "../../../../components/config_section";

export const SourceConfigSection = ({
  job: {
    job_config: {
      spec: { source },
    },
  },
}) => {
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
        <ConfigSectionPanel title="Join Columns">
          <EuiListGroup
            flush
            listItems={source.join_on.map((item) => ({ label: item }))}
          />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

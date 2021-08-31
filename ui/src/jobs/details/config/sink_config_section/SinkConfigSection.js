import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../components/config_section";
import { BigQuerySinkConfigTable } from "./bq_config_section/BigQuerySinkConfigTable";
import { SinkConfigTable } from "./sink_config_table/SinkConfigTable";
import { SinkType } from "../../../../services/job/SinkType";

export const SinkConfigSection = ({
  job: {
    job_config: {
      spec: { ensembler, sink },
    },
  },
}) => {
  return (
    <EuiFlexGroup direction="row">
      <EuiFlexItem grow={2}>
        <ConfigSectionPanel title="Sink">
          <SinkConfigTable sink={sink} ensembler={ensembler} />
        </ConfigSectionPanel>
      </EuiFlexItem>

      {
        /* eslint-disable-next-line eqeqeq */
        sink.type == SinkType.BQ && (
          <EuiFlexItem grow={3}>
            <ConfigSectionPanel title="BigQuery Configuration">
              <BigQuerySinkConfigTable bqConfig={sink.bq_config} />
            </ConfigSectionPanel>
          </EuiFlexItem>
        )
      }
    </EuiFlexGroup>
  );
};

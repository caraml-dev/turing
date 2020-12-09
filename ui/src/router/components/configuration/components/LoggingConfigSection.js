import React, { Fragment } from "react";
import { ConfigSectionPanel } from "./section";
import {
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
  EuiPanel
} from "@elastic/eui";
import { getBigQueryConsoleUrl } from "../../../../utils/bq";

const BigQueryConfigTable = ({
  bigquery_config: { table, service_account_secret }
}) => {
  const items = [
    {
      title: "BigQuery Table",
      description: (
        <EuiLink href={getBigQueryConsoleUrl(table)} target="_blank" external>
          {table}
        </EuiLink>
      )
    },
    {
      title: "Service Account",
      description: service_account_secret
    }
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "25%" } }}
      descriptionProps={{ style: { width: "75%" } }}
    />
  );
};

const KafkaConfigTable = ({
  kafka_config: { brokers, topic, serialization_format }
}) => {
  const items = [
    {
      title: "Broker(s)",
      description: brokers
    },
    {
      title: "Topic",
      description: topic
    },
    {
      title: "Serialization Format",
      description: serialization_format
    }
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "25%" } }}
      descriptionProps={{ style: { width: "75%" } }}
    />
  );
};

const BigQueryConfigSection = ({ bigquery_config }) => {
  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel title="Results Logs">
          Google BigQuery
        </ConfigSectionPanel>
      </EuiFlexItem>
      <EuiFlexItem grow={2}>
        <ConfigSectionPanel title="BigQuery Configuration">
          <BigQueryConfigTable bigquery_config={bigquery_config} />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

const KafkaConfigSection = ({ kafka_config }) => {
  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel title="Results Logs">Kafka</ConfigSectionPanel>
      </EuiFlexItem>
      <EuiFlexItem grow={2}>
        <ConfigSectionPanel title="Kafka Configuration">
          <KafkaConfigTable kafka_config={kafka_config} />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

export const LoggingConfigSection = ({ config: { log_config } }) => {
  const { result_logger_type } = log_config;

  return (
    <Fragment>
      {result_logger_type === "bigquery" && (
        <BigQueryConfigSection bigquery_config={log_config.bigquery_config} />
      )}

      {result_logger_type === "kafka" && (
        <KafkaConfigSection kafka_config={log_config.kafka_config} />
      )}

      {result_logger_type === "nop" && <EuiPanel>Not Configured</EuiPanel>}
    </Fragment>
  );
};

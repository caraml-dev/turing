import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ResultLoggingTypePanel } from "../components/outcome_config/ResultLoggingTypePanel";
import { FormContext } from "../../../../components/form/context";
import { get } from "../../../../components/form/utils";
import { BigQueryConfigPanel } from "../components/outcome_config/bigquery/BigQueryConfigPanel";
import { KafkaConfigPanel } from "../components/outcome_config/kafka/KafkaConfigPanel";
import { SecretsContextProvider } from "../../../../providers/secrets/context";
import FormValidationContext from "../../../../components/form/validation";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import { resultLoggingOptions } from "../components/outcome_config/typeOptions";

export const OutcomeStep = ({ projectId }) => {
  const {
    data: {
      config: { log_config },
    },
    onChangeHandler,
  } = useContext(FormContext);

  const { errors } = useContext(FormValidationContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <ResultLoggingTypePanel
          type={log_config.result_logger_type}
          options={resultLoggingOptions}
          onChange={onChange("config.log_config.result_logger_type")}
          errors={get(errors, "config.log_config.result_logger_type")}
        />
      </EuiFlexItem>

      {log_config.result_logger_type === "bigquery" && (
        <EuiFlexItem>
          <SecretsContextProvider projectId={projectId}>
            <BigQueryConfigPanel
              projectId={projectId}
              bigQueryConfig={log_config.bigquery_config}
              onChangeHandler={onChange("config.log_config.bigquery_config")}
              errors={get(errors, "config.log_config.bigquery_config")}
            />
          </SecretsContextProvider>
        </EuiFlexItem>
      )}

      {log_config.result_logger_type === "kafka" && (
        <EuiFlexItem>
          <KafkaConfigPanel
            kafkaConfig={log_config.kafka_config}
            onChangeHandler={onChange("config.log_config.kafka_config")}
            errors={get(errors, "config.log_config.kafka_config")}
          />
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};

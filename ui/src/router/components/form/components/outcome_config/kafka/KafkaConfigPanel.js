import React from "react";
import { Panel } from "../../Panel";
import {
  EuiFieldText,
  EuiForm,
  EuiFormRow,
  EuiLink,
  EuiSpacer,
  EuiSuperSelect
} from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";
import { resultLoggingConfig } from "../../../../../../config.js";

const logSerializationFormatOptions = [
  {
    value: "json",
    inputDisplay: "JSON"
  },
  {
    value: "protobuf",
    inputDisplay: "Protobuf",
    helpText: !!resultLoggingConfig.protoUrl ? (
      <EuiLink href={resultLoggingConfig.protoUrl} target="_blank" external>
        TuringResultLog.proto
      </EuiLink>
    ) : (
      ""
    )
  }
];

export const KafkaConfigPanel = ({
  kafkaConfig,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const selectedFormat = logSerializationFormatOptions.find(
    option => option.value === kafkaConfig.serialization_format
  );

  return (
    <Panel title="Kafka Configuration">
      <EuiSpacer size="m" />
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Brokers *"
              content="A comma-separated list of host and port pairs that make up the Kafka broker address(es)"
            />
          }
          isInvalid={!!errors.brokers}
          error={errors.brokers}
          display="row">
          <EuiFieldText
            fullWidth
            placeholder="host:port"
            value={kafkaConfig.brokers || ""}
            onChange={e => onChange("brokers")(e.target.value)}
            isInvalid={!!errors.brokers}
            name="kafka-brokers"
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFormRow
          fullWidth
          label="Topic *"
          isInvalid={!!errors.topic}
          error={errors.topic}>
          <EuiFieldText
            fullWidth
            value={kafkaConfig.topic || ""}
            onChange={e => onChange("topic")(e.target.value)}
            isInvalid={!!errors.topic}
            name="kafka-topic"
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Serialization Format *"
              content="Select the message serialization format to be used when writing to the Kafka topic"
            />
          }
          isInvalid={!!errors.serialization_format}
          error={errors.serialization_format}
          helpText={(selectedFormat || {}).helpText || ""}
          display="row">
          <EuiSuperSelect
            fullWidth
            options={logSerializationFormatOptions}
            valueOfSelected={(selectedFormat || {}).value || ""}
            onChange={onChange("serialization_format")}
            isInvalid={!!errors.serialization_format}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};

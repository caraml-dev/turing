import React from "react";
import { Panel } from "../../Panel";
import { EuiFieldText, EuiForm, EuiFormRow, EuiSpacer } from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";

export const KafkaConfigPanel = ({
  kafkaConfig,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

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
      </EuiForm>
    </Panel>
  );
};

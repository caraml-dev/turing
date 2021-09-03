import React from "react";
import { EuiDescriptionList, EuiText } from "@elastic/eui";

export const BigQueryOptionsTable = ({ options }) =>
  options && Object.entries(options).length ? (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={Object.entries(options).map(([title, description]) => ({
        title,
        description,
      }))}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  ) : (
    <EuiText color="subdued" size="s">
      None
    </EuiText>
  );

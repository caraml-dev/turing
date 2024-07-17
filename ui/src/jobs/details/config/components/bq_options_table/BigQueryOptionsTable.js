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
      columnWidths={[1, 7/3]}
    />
  ) : (
    <EuiText color="subdued" size="s">
      None
    </EuiText>
  );

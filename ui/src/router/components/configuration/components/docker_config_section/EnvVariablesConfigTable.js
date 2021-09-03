import React from "react";
import { EuiDescriptionList, EuiText } from "@elastic/eui";
import { ExpandableContainer } from "../../../../../components/expandable_container/ExpandableContainer";

export const EnvVariablesConfigTable = ({ variables }) => {
  const items = variables.map((v) => ({
    title: v.name,
    description: v.value,
  }));

  return variables.length ? (
    <ExpandableContainer maxCollapsedHeight={300}>
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        titleProps={{ style: { width: "30%" } }}
        descriptionProps={{ style: { width: "70%" } }}
      />
    </ExpandableContainer>
  ) : (
    <EuiText size="s" color="subdued">
      None
    </EuiText>
  );
};

import React from "react";
import { EuiInMemoryTable, EuiText } from "@elastic/eui";
import { ExpandableContainer } from "../../../../../components/expandable_container/ExpandableContainer";
import "./SecretsConfigTable.scss"

export const SecretsConfigTable = ({ variables }) => {
  const columns = [
    {
      field: "mlp_secret_name",
      name: "MLP Secret Name",
      width: "30%",
      sortable: true
    },
    {
      field: "env_var_name",
      name: "Environment Variable Name",
      width: "70%",
      sortable: true
    }
  ];

  return variables.length ? (
    <ExpandableContainer maxCollapsedHeight={300}>
      <EuiInMemoryTable
        className={"euiInMemoryTable--secretsConfigTable"}
        items={variables}
        columns={columns}
        itemId="name"
      />
    </ExpandableContainer>
  ) : (
    <EuiText size="s" color="subdued">
      None
    </EuiText>
  );
};

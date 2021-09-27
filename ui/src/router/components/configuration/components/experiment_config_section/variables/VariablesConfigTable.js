import React, { Fragment, useMemo } from "react";
import {
  EuiDescriptionList,
  EuiHorizontalRule,
  EuiText,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import orderBy from "lodash/orderBy";
import "./VariablesConfigTable.scss";

const makeListItems = (source, value) => {
  const fieldValue = !!value ? value : "-";
  return [
    {
      title: "Field Source",
      description: source,
    },
    {
      title: "Value",
      description: <span title={fieldValue}>{fieldValue}</span>,
    },
  ];
};

export const VariablesConfigTable = ({ variables }) => {
  const variablesList = useMemo(() => {
    return orderBy(
      variables.config,
      ["required", "name"],
      ["desc", "asc"]
    ).filter((v) => !!v.field);
  }, [variables.config]);
  return !!variables.config.length ? (
    <Fragment>
      {variablesList.map(({ name, field_source, field }, idx) => (
        <Fragment key={name}>
          <EuiTitle size="xxs">
            <EuiTextColor color="secondary">{name}</EuiTextColor>
          </EuiTitle>
          <EuiDescriptionList
            compressed
            textStyle="reverse"
            type="responsiveColumn"
            listItems={makeListItems(field_source, field)}
          />
          {idx < variables.config.length - 1 && (
            <EuiHorizontalRule size="full" margin="xs" />
          )}
        </Fragment>
      ))}
    </Fragment>
  ) : (
    <EuiText size="s" color="subdued">
      None
    </EuiText>
  );
};

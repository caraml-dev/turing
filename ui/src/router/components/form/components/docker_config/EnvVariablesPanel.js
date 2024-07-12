import React from "react";
import { EuiButtonIcon, EuiFieldText, EuiSpacer } from "@elastic/eui";
import { Panel } from "../Panel";
import { InMemoryTableForm } from "../../../../../components/form/in_memory_table_form/InMemoryTableForm";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";

export const EnvVariablesPanel = ({
  variables,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const items = [
    ...variables.map((v, idx) => ({ idx, ...v })),
    { idx: variables.length },
  ];

  const onDeleteVariable = (idx) => () => {
    variables.splice(idx, 1);
    onChangeHandler(variables);
  };

  const getRowProps = (item) => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`,
    };
  };

  const columns = [
    {
      name: "Name",
      field: "name",
      width: "45%",
      render: (name, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Name"
          value={name || ""}
          onChange={(e) => onChange(`${item.idx}.name`)(e.target.value)}
        />
      ),
    },
    {
      name: "Value",
      field: "value",
      width: "45%",
      render: (value, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Value"
          value={value || ""}
          onChange={(e) => onChange(`${item.idx}.value`)(e.target.value)}
        />
      ),
    },
    {
      width: "10%",
      actions: [
        {
          render: (item) =>
            item.idx < items.length - 1 ? (
              <EuiButtonIcon
                size="s"
                color="danger"
                iconType="trash"
                onClick={onDeleteVariable(item.idx)}
                aria-label="Remove variable"
              />
            ) : (
              <div />
            ),
        },
      ],
    },
  ];

  return (
    <Panel title="Environment Variables">
      <EuiSpacer size="xs" />
      <InMemoryTableForm
        columns={columns}
        rowProps={getRowProps}
        items={items}
        errors={errors}
        renderErrorHeader={(key) => `Row ${parseInt(key) + 1}`}
      />
    </Panel>
  );
};

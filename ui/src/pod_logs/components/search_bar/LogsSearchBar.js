import React, { useMemo } from "react";
import { EuiSearchBar } from "@elastic/eui";

export const LogsSearchBar = ({ components, params, onParamsChange }) => {
  const filters = useMemo(() => {
    return [
      {
        type: "field_value_toggle_group",
        field: "component_type",
        items: components,
      },
      {
        type: "field_value_selection",
        field: "tail_lines",
        name: "Log Tail",
        multiSelect: false,
        options: [
          {
            value: "100",
            name: "Last 100 records",
          },
          {
            value: "1000",
            name: "Last 1000 records",
          },
          {
            value: "",
            name: "From the container start",
          },
        ],
      },
    ];
  }, [components]);

  const queryString = useMemo(() => {
    return Object.entries(params)
      .map(([k, v]) => `${k}:"${v}"`)
      .join(" ");
  }, [params]);

  const onChange = ({ query, error }) => {
    if (!error) {
      const newParams = {
        ...params,
        ...query.ast.clauses.reduce((acc, { field, value }) => {
          acc[field] = value;
          return acc;
        }, {}),
      };

      if (JSON.stringify(newParams) !== JSON.stringify(params)) {
        onParamsChange(newParams);
      }
    }
  };

  return (
    <EuiSearchBar
      query={queryString}
      box={{ readOnly: true }}
      filters={filters}
      onChange={onChange}
    />
  );
};

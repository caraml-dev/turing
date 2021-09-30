import React, { useEffect, useMemo, useState } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiSearchBar,
  EuiSpacer,
} from "@elastic/eui";
import { LazyLog, ScrollFollow } from "react-lazylog";
import { slugify } from "@gojek/mlp-ui";

import "./PodLogsViewer.scss";

export const PodLogsViewer = ({
  components,
  emitter,
  query,
  onQueryChange,
  batchSize,
}) => {
  const filters = useMemo(
    () => [
      {
        type: "field_value_toggle_group",
        field: "component_type",
        items: components,
      },
      {
        type: "field_value_selection",
        field: "tail_lines",
        name: !!query.tail_lines
          ? `Last ${query.tail_lines} records`
          : "From the container start",
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
            value: "all",
            name: "From the container start",
          },
        ],
      },
    ],
    [components, query.tail_lines]
  );

  const [filterValues, setFilterValues] = useState({});

  useEffect(() => {
    const filterValues = Object.fromEntries([
      ...Object.entries(query || {}).filter(([k]) =>
        filters.some((f) => f.field === k)
      ),
      ["tail_lines", query.tail_lines || "all"],
    ]);

    setFilterValues((old) =>
      JSON.stringify(old) !== JSON.stringify(filterValues) ? filterValues : old
    );
  }, [query, filters, components, setFilterValues]);

  const searchQuery = useMemo(() => {
    return Object.entries(filterValues)
      .map(([k, v]) => `${k}:"${v}"`)
      .join(" ");
  }, [filterValues]);

  const onChange = ({ query: { ast }, error }) => {
    if (!error) {
      let newFilterValues = {
        ...filterValues,
        ...ast.clauses.reduce((acc, { field, value }) => {
          acc[field] = value;
          return acc;
        }, {}),
      };

      if (JSON.stringify(newFilterValues) !== JSON.stringify(filterValues)) {
        if (newFilterValues.tail_lines === "all") {
          delete newFilterValues["tail_lines"];
          newFilterValues.head_lines = batchSize;
        }

        onQueryChange(() => newFilterValues);
      }
    }
  };

  const search = {
    query: searchQuery,
    box: {
      readOnly: true,
    },
    filters,
    onChange,
  };

  return (
    <EuiFlexGroup
      direction="column"
      gutterSize="none"
      className="euiFlexGroup---logsContainer">
      <EuiFlexItem grow={false}>
        <EuiSearchBar {...search} />
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        <EuiSpacer size="s" />
      </EuiFlexItem>
      <EuiFlexItem grow={true}>
        <ScrollFollow
          startFollowing={true}
          render={({ onScroll, follow }) => (
            <LazyLog
              key={slugify(searchQuery)}
              eventSource={emitter}
              extraLines={1}
              onScroll={onScroll}
              follow={follow}
              caseInsensitive
              enableSearch
              selectableLines
            />
          )}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

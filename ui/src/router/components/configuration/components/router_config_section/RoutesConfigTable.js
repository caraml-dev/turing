import React, { useMemo, useState } from "react";
import {
  EuiButtonIcon,
  EuiCopy,
  EuiIcon,
  EuiInMemoryTable,
  EuiLink,
  EuiToolTip
} from "@elastic/eui";
import { RoutesTableExpandedRow } from "./RoutesTableExpandedRow";

import "./RoutesConfigTable.scss";

export const RoutesConfigTable = ({ routes, rules = [], defaultRouteId }) => {
  const [itemIdToExpandedRowMap, setItemIdToExpandedRowMap] = useState({});

  const toggleDetails = item => () => {
    const itemIdToExpandedRowMapValues = { ...itemIdToExpandedRowMap };
    if (itemIdToExpandedRowMapValues[item.id]) {
      delete itemIdToExpandedRowMapValues[item.id];
    } else {
      itemIdToExpandedRowMapValues[item.id] = (
        <RoutesTableExpandedRow item={item} />
      );
    }
    setItemIdToExpandedRowMap(itemIdToExpandedRowMapValues);
  };

  const routesWithRules = useMemo(() => {
    const routeRules = rules.reduce((acc, rule) => {
      rule.routes.forEach(route => {
        !!acc[route]
          ? (acc[route] = [...acc[route], rule.conditions])
          : (acc[route] = [rule.conditions]);
      });
      return acc;
    }, {});

    return routes
      .sort((r1, r2) =>
        r1.id === defaultRouteId ? -1 : r2.id === defaultRouteId ? 1 : 0
      )
      .map(r => ({
        ...r,
        rules: routeRules[r.id]
      }));
  }, [routes, defaultRouteId, rules]);

  const columns = [
    {
      width: "24px",
      render: () => (
        <EuiIcon type="graphApp" size="m" style={{ verticalAlign: "sub" }} />
      )
    },
    {
      field: "id",
      width: "15%",
      name: "Id",
      render: (id, item) =>
        item.endpoint_info_url ? (
          <EuiToolTip content="Open endpoint details page">
            <EuiLink href={item.endpoint_info_url}>{id}</EuiLink>
          </EuiToolTip>
        ) : (
          id
        )
    },
    {
      field: "endpoint",
      width: "70%",
      name: "Endpoint",
      truncateText: true,
      render: endpoint => (
        <EuiCopy
          textToCopy={endpoint}
          beforeMessage="Click to copy URL to clipboard">
          {copy => (
            <EuiLink onClick={copy}>
              <EuiIcon type={"copyClipboard"} size="s" />
              {endpoint}
            </EuiLink>
          )}
        </EuiCopy>
      )
    },
    {
      field: "timeout",
      width: "10%",
      name: "Timeout"
    },
    {
      width: "5%",
      actions: [
        {
          render: item =>
            !!item.rules ? (
              <EuiToolTip content="Show traffic rules">
                <EuiButtonIcon
                  size="s"
                  className="euiIconButton--action"
                  iconType="logstashIf"
                  onClick={toggleDetails(item)}
                  aria-label="Remove data field"
                />
              </EuiToolTip>
            ) : (
              <div />
            )
        }
      ]
    }
  ];

  const getRowProps = item => {
    const { id } = item;
    return id === defaultRouteId
      ? {
          className: "euiTableRow-isSelected",
          title: "Default Route"
        }
      : {};
  };

  const getCellProps = (item, column) => {
    const { id } = item;
    const { field } = column;
    return {
      "data-test-subj": `cell-${id}-${field}`,
      textOnly: true
    };
  };

  return (
    <EuiInMemoryTable
      items={routesWithRules}
      columns={columns}
      itemId="id"
      itemIdToExpandedRowMap={itemIdToExpandedRowMap}
      isSelectable={false}
      rowProps={getRowProps}
      cellProps={getCellProps}
    />
  );
};

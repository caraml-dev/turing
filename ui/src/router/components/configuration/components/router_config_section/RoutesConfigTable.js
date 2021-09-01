import React, { useContext, useMemo, useState } from "react";
import {
  EuiButtonIcon,
  EuiCopy,
  EuiIcon,
  EuiInMemoryTable,
  EuiLink,
  EuiToolTip,
} from "@elastic/eui";
import { CurrentProjectContext } from "@gojek/mlp-ui";
import { RoutesTableExpandedRow } from "./RoutesTableExpandedRow";
import { ANNOTATIONS_MERLIN_MODEL_ID } from "../../../../../providers/endpoints/MerlinEndpointsProvider";

const MerlinRouteId = ({ modelId, routeId }) => {
  const { projectId } = useContext(CurrentProjectContext);

  return (
    <EuiToolTip content="Open endpoint details page">
      <EuiLink href={`/merlin/projects/${projectId}/models/${modelId}/`}>
        {routeId}
      </EuiLink>
    </EuiToolTip>
  );
};

const RouteId = ({ route }) => {
  return route.annotations && route.annotations[ANNOTATIONS_MERLIN_MODEL_ID] ? (
    <MerlinRouteId
      routeId={route.id}
      modelId={route.annotations[ANNOTATIONS_MERLIN_MODEL_ID]}
    />
  ) : (
    route.id
  );
};

export const RoutesConfigTable = ({ routes, rules = [], defaultRouteId }) => {
  const [itemIdToExpandedRowMap, setItemIdToExpandedRowMap] = useState({});

  const toggleDetails = (item) => () => {
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
      rule.routes.forEach((route) => {
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
      .map((r) => ({
        ...r,
        rules: routeRules[r.id],
      }));
  }, [routes, defaultRouteId, rules]);

  const columns = [
    {
      width: "24px",
      render: () => (
        <EuiIcon type="graphApp" size="m" style={{ verticalAlign: "sub" }} />
      ),
    },
    {
      field: "id",
      width: "15%",
      name: "Id",
      render: (_, item) => <RouteId route={item} />,
    },
    {
      field: "endpoint",
      width: "70%",
      name: "Endpoint",
      truncateText: true,
      render: (endpoint) => (
        <EuiCopy
          textToCopy={endpoint}
          beforeMessage="Click to copy URL to clipboard">
          {(copy) => (
            <EuiLink onClick={copy}>
              <EuiIcon type={"copyClipboard"} size="s" />
              &nbsp;
              {endpoint}
            </EuiLink>
          )}
        </EuiCopy>
      ),
    },
    {
      field: "timeout",
      width: "10%",
      name: "Timeout",
    },
    {
      width: "5%",
      actions: [
        {
          render: (item) =>
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
            ),
        },
      ],
    },
  ];

  const getRowProps = (item) => {
    const { id } = item;
    return id === defaultRouteId
      ? {
          className: "euiTableRow-isSelected",
          title: "Default Route",
        }
      : {};
  };

  const getCellProps = (item, column) => {
    const { id } = item;
    const { field } = column;
    return {
      "data-test-subj": `cell-${id}-${field}`,
      textOnly: true,
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

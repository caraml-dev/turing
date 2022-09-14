import React, { Fragment, useMemo } from "react";
import {
  EuiBadge,
  EuiIcon,
  EuiInMemoryTable,
  EuiSpacer,
} from "@elastic/eui";
import { RouteId } from "./RoutesConfigTable";

export const RulesConfigTable = ({ routes, rules = [], defaultTrafficRule }) => {
  const rulesWithRouteDetails = useMemo(() => {
    const details = [];
    rules.forEach((rule) => {
        const ruleRoutes = [];
        rule.routes.forEach((ruleRoute) => {
            const filteredRoute = routes.find(route => route.id === ruleRoute);
            ruleRoutes.push(filteredRoute);
        })
        details.push(ruleRoutes);
    });
    const updatedRules = rules.map(function(rule, idx) {
        return {...rule, routes: details[idx]};
    });

    // Append Default Traffic Rule
    const defaultRuleRoutes = [];
    defaultTrafficRule.routes.forEach((ruleRoute) => {
        const filteredRoute = routes.find(route => route.id === ruleRoute);
        defaultRuleRoutes.push(filteredRoute);
    });
    updatedRules.push({
      name: "default-traffic-rule",
      conditions: [],
      routes: defaultRuleRoutes
    });
    return updatedRules;
  }, [routes, rules, defaultTrafficRule]);

  const columns = [
    {
      width: "24px",
      render: () => (
        <EuiIcon type="logstashIf" size="m" style={{ verticalAlign: "sub" }} />
      ),
    },
    {
      field: "name",
      width: "10%",
      name: "Name",
    },
    {
      field: "conditions",
      width: "60%",
      name: "Conditions",
      truncateText: true,
      render: (conditions) => {
        return conditions.map((condition, idx) => (
            <Fragment key={idx}>
              <EuiBadge key={`field-source-${idx}`} color="primary">
                {condition.field_source}
              </EuiBadge>
              <EuiBadge key={`field-${idx}`}>
                {condition.field}
              </EuiBadge>
              <EuiBadge key={`operator-${idx}`} color="warning">
                {condition.operator}
              </EuiBadge>
              {condition.values.map((val, idx) => (
                <EuiBadge key={`filter-${idx}`}>
                  {val}
                </EuiBadge>
              ))}
              <EuiSpacer size="xs" />
            </Fragment>
        ))
      },
    },
    {
        field: "routes",
        width: "20%",
        name: "Routes",
        truncateText: true,
        render: (routes) => {
          return routes.map((route, idx) => (
            <Fragment key={idx}>
                <RouteId route={route} />
                <EuiSpacer size="xs" />
            </Fragment>
          ))
        },
    },
    {
        field: "routes",
        width: "10%",
        name: "Timeout",
        truncateText: true,
        render: (routes) => {
          return routes.map((route, idx) => (
            <Fragment key={idx}>
                {route.timeout}
                <EuiSpacer size="xs" />
            </Fragment>
          ))
        },
    },
  ];

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
      items={rulesWithRouteDetails}
      columns={columns}
      itemId="id"
      isSelectable={false}
      cellProps={getCellProps}
    />
  );
};

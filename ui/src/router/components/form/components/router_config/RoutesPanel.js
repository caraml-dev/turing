import React, { useContext } from "react";
import { Panel } from "../Panel";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { RouteCard } from "./route_card/RouteCard";
import { newRoute } from "../../../../../services/router/TuringRouter";
import { get } from "../../../../../components/form/utils";
import EndpointsContext from "../../../../../providers/endpoints/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";

export const RoutesPanel = ({
  routes,
  protocol,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const endpoints = useContext(EndpointsContext);
  const filteredEndpoints = endpoints.map((endpoint) => {
    return {
      ...endpoint,
      options: endpoint.options.filter((option) => {
        if (protocol === "UPI_V1") {
          return !option.label.startsWith("http");
        }
        return option.label.startsWith("http");
      }),
    };
  });

  const onAddRoute = () => {
    onChange("routes")([...routes, newRoute(protocol)]);
  };

  const onDeleteRoute = (idx) => () => {
    routes.splice(idx, 1);
    onChange("routes")(routes);
  };

  return (
    <Panel title="Routes">
      <EuiFlexGroup direction="column" gutterSize="s">
        {routes.map((route, idx) => (
          <EuiFlexItem key={`route-${idx}`}>
            <RouteCard
              route={route}
              protocol={protocol}
              endpointOptions={filteredEndpoints}
              onChange={onChange(`routes.${idx}`)}
              onDelete={routes.length > 1 ? onDeleteRoute(idx) : null}
              errors={get(errors, `${idx}`)}
            />
            <EuiSpacer size="s" />
          </EuiFlexItem>
        ))}
        <EuiFlexItem>
          <EuiButton onClick={onAddRoute}>+ Add Route</EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};

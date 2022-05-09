import React, { useContext } from "react";
import { Panel } from "../Panel";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { RouteCard } from "./route_card/RouteCard";
import { newRoute } from "../../../../../services/router/TuringRouter";
import { get } from "../../../../../components/form/utils";
import EndpointsContext from "../../../../../providers/endpoints/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";

export const RoutesPanel = ({ routes, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const endpoints = useContext(EndpointsContext);

  const onAddRoute = () => {
    onChange("routes")([...routes, newRoute()]);
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
              endpointOptions={endpoints}
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

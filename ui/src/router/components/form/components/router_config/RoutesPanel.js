import React, { useContext, useEffect, useState } from "react";
import { Panel } from "../Panel";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { RouteCard } from "./route_card/RouteCard";
import { newRoute } from "../../../../../services/router/TuringRouter";
import { get } from "../../../../../components/form/utils";
import EndpointsContext from "../../../../../providers/endpoints/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";

export const RoutesPanel = ({
  routes,
  defaultRouteId,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const [defaultRouteIndex, setDefaultRouteIndex] = useState(
    (() => {
      let defaultRouteIdx = routes.findIndex(r => r.id === defaultRouteId);
      if (defaultRouteIdx < 0) {
        defaultRouteIdx = 0;
      }
      return defaultRouteIdx;
    })()
  );

  const endpoints = useContext(EndpointsContext);

  const onAddRoute = () => {
    onChange("routes")([...routes, newRoute()]);
  };

  const onDeleteRoute = idx => () => {
    routes.splice(idx, 1);
    onChange("routes")(routes);
    idx < defaultRouteIndex && setDefaultRouteIndex(idx => idx - 1);
  };

  const defaultRouteIdUpdated = routes[defaultRouteIndex].id;

  useEffect(() => {
    onChange("default_route_id")(defaultRouteIdUpdated);
  }, [defaultRouteIdUpdated, onChange]);

  return (
    <Panel title="Routes">
      <EuiFlexGroup direction="column" gutterSize="s">
        {routes.map((route, idx) => (
          <EuiFlexItem key={`route-${idx}`}>
            <RouteCard
              route={route}
              isDefault={idx === defaultRouteIndex}
              endpointOptions={endpoints}
              onChange={onChange(`routes.${idx}`)}
              onSelect={() => setDefaultRouteIndex(idx)}
              onDelete={onDeleteRoute(idx)}
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

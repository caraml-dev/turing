import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { TreatmentMappingConfigSection } from "./TreatmentMappingConfigSection";
import { FallbackRouteConfigSection } from "./FallbackRouteConfigSection";
import { RoutingOrderConfigSection } from "./RoutingOrderConfigSection";
import ExperimentEngineContext from "../../../../../providers/experiments/context";
import { Panel } from "../../../form/components/Panel";
import ExperimentEngineComponentLoader from "../../../../../components/remote_component/ExperimentEngineComponentLoader";
import { RemoteComponent } from "../../../../../components/remote_component/RemoteComponent";

const FallbackView = ({ text }) => (
  <EuiFlexItem grow={true}>
    <Panel title="Route Selection">
      <EuiSpacer size="m" />
      {text}
    </Panel>
  </EuiFlexItem>
);

const StandardEnsemblerWithCustomExperimentEngineConfigView = ({
  remoteUi,
  projectId,
  routes,
  routeNamePath
}) => {
  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Standard Ensembler config for the selected Custom Experiment Engine" />}>
      <ExperimentEngineComponentLoader
        FallbackView={FallbackView}
        remoteUi={remoteUi}
        componentName="Standard Ensembler"
      >
        <RemoteComponent
          scope={remoteUi.name}
          name="./StandardEnsemblerConfigDetails"
          fallback={<FallbackView text="Loading Standard Ensembler config for the selected Custom Experiment Engine" />}
          projectId={projectId}
          routes={routes}
          routeNamePath={routeNamePath}
        />
      </ExperimentEngineComponentLoader>
    </React.Suspense>
  );
};

export const StandardConfigViewGroup = ({
  projectId,
  routes,
  standardConfig,
  experimentConfig,
  type
}) => {
  const { getEngineProperties, isLoaded } = useContext(ExperimentEngineContext);

  const engineProps = getEngineProperties(type);

  return (
    !!standardConfig && (
      <EuiFlexGroup direction="column">
        {engineProps?.type === "standard" ? (
          <EuiFlexItem>
            <TreatmentMappingConfigSection
              engine={type}
              experiments={experimentConfig}
              mappings={standardConfig.experiment_mappings}
            />
          </EuiFlexItem>
        ) : isLoaded ? (
          <EuiFlexItem>
            <StandardEnsemblerWithCustomExperimentEngineConfigView
              remoteUi={engineProps.custom_experiment_manager_config.remote_ui}
              projectId={projectId}
              routes={routes}
              routeNamePath={standardConfig.route_name_path}
            />
          </EuiFlexItem>
        ) : (
          <EuiFlexItem>
            <FallbackView text={"Loading ..."} />
          </EuiFlexItem>
        )}

        <EuiFlexItem>
          <EuiFlexGroup direction="row">
            <EuiFlexItem grow={1}>
              <RoutingOrderConfigSection
                isLazyRouting={!!standardConfig.lazy_routing}
              />
            </EuiFlexItem>
            <EuiFlexItem grow={2}>
              <FallbackRouteConfigSection
                fallbackResponseRouteId={standardConfig.fallback_response_route_id}
              />
            </EuiFlexItem>
          </EuiFlexGroup>
        </EuiFlexItem>
      </EuiFlexGroup>
    )
  );
};
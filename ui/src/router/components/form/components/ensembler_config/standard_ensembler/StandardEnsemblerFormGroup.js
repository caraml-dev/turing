import React, { useContext, useEffect } from "react";
import { EuiFlexItem, EuiSpacer, EuiText } from "@elastic/eui";

import { get } from "../../../../../../components/form/utils";
import { StandardEnsembler } from "../../../../../../services/ensembler";
import { StandardEnsemblerPanel } from "./StandardEnsemblerPanel";
import { RouteSelectionPanel } from "../RouteSelectionPanel";
import { FormLabelWithToolTip } from "../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";
import ExperimentEngineContext from "../../../../../../providers/experiments/context";
import { Panel } from "../../Panel";
import { RemoteComponent } from "../../../../../../components/remote_component/RemoteComponent";
import RemoteLoaderComponent from "../../../../../../components/remote_component/RemoteLoaderComponent";

const FallbackView = ({ text }) => (
  <EuiFlexItem grow={true}>
    <Panel title="Route Selection">
      <EuiSpacer size="m" />
      {text}
    </Panel>
  </EuiFlexItem>
);

const StandardEnsemblerWithCustomExperimentEnginePanel = ({
  remoteUi,
  projectId,
  routes,
  routeNamePath,
  onChange,
  errors,
}) => {
  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Standard Ensembler config for the selected Custom Experiment Engine" />}>
      <RemoteLoaderComponent
        FallbackView={FallbackView}
        remoteUi={remoteUi}
        componentName="Standard Ensembler"
      >
        <RemoteComponent
          scope={remoteUi.name}
          name="./EditStandardEnsemblerConfig"
          fallback={<FallbackView text="Loading Standard Ensembler form for the selected Custom Experiment Engine" />}
          projectId={projectId}
          routes={routes}
          routeNamePath={routeNamePath}
          onChange={onChange}
          errors={errors}
        />
      </RemoteLoaderComponent>
    </React.Suspense>
  );
};

export const StandardEnsemblerFormGroup = ({
  projectId,
  experimentEngine = {},
  routes,
  rules,
  default_traffic_rule,
  standardConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const { getEngineProperties, isLoaded } = useContext(ExperimentEngineContext);

  useEffect(() => {
    !standardConfig && onChangeHandler(StandardEnsembler.newConfig());
  }, [standardConfig, onChangeHandler]);

  const engineProps = getEngineProperties(experimentEngine.type);

  const routeOptions = [
    {
      value: "nop",
      inputDisplay: "Select a route...",
      disabled: true,
    },
    ...routes.map((route) => ({
      icon: "graphApp",
      value: route.id,
      inputDisplay: route.id,
      dropdownDisplay: (
        <>
          <strong>{route.id}</strong>
          <EuiText color="subdued" size="s">
            {route.endpoint}
          </EuiText>
        </>
      ),
    })),
  ];

  return (
    !!standardConfig && (
      <>
        {engineProps?.type === "standard" ? (
          <EuiFlexItem>
            <StandardEnsemblerPanel
              experiments={experimentEngine.config.experiments}
              mappings={standardConfig.experiment_mappings}
              routeOptions={routeOptions}
              onChangeHandler={onChange("experiment_mappings")}
              errors={get(errors, "experiment_mappings")}
            />
          </EuiFlexItem>
        ) : isLoaded ? (
          <EuiFlexItem>
            <StandardEnsemblerWithCustomExperimentEnginePanel
              remoteUi={engineProps.custom_experiment_manager_config.remote_ui}
              projectId={projectId}
              routes={routes}
              routeNamePath={standardConfig.route_name_path}
              onChange={onChange("route_name_path")}
              errors={get(errors, "route_name_path")}
            />
          </EuiFlexItem>
        ) : (
          <EuiFlexItem>
            <FallbackView text={"Loading ..."} />
          </EuiFlexItem>
        )}

        <EuiFlexItem>
          <RouteSelectionPanel
            routeId={standardConfig.fallback_response_route_id}
            routes={routes}
            rules={rules}
            default_traffic_rule={default_traffic_rule}
            onChange={onChange("fallback_response_route_id")}
            errors={get(errors, "fallback_response_route_id")}
            panelTitle="Fallback"
            inputLabel={
              <FormLabelWithToolTip
                label="Fallback Response *"
                content="Select the route to fallback to, if the call to the experiment engine fails or if a matching route cannot be found for the treatment."
              />
            }
          />
        </EuiFlexItem>
      </>
    )
  );
};

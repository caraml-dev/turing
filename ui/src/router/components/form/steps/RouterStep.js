import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { GeneralSettingsPanel } from "../components/router_config/GeneralSettingsPanel";
import { RoutesPanel } from "../components/router_config/RoutesPanel";
import { FormContext, FormValidationContext } from "@gojek/mlp-ui";
import { MerlinEndpointsProvider } from "../../../../providers/endpoints/MerlinEndpointsProvider";
import { useConfig } from "../../../../config";
import { get } from "../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import { RulesPanel } from "../components/router_config/RulesPanel";
import { ResourcesRequirementsPanel } from "../components/ResourcesRequirementsPanel";

export const RouterStep = ({ projectId }) => {
  const {
    appConfig: {
      scaling: { maxAllowedReplica },
    },
  } = useConfig();
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <GeneralSettingsPanel
          name={data.name}
          environment={data.environment_name}
          timeout={data.config.timeout}
          protocol={data.config.protocol}
          isEdit={!!data.id}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        <MerlinEndpointsProvider
          projectId={projectId}
          environmentName={data.environment_name}
        >
          <RoutesPanel
            protocol={data.config.protocol}
            routes={get(data, "config.routes")}
            onChangeHandler={onChange("config")}
            errors={get(errors, "config.routes")}
          />
        </MerlinEndpointsProvider>
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        <RulesPanel
          default_traffic_rule={get(data, "config.default_traffic_rule")}
          default_traffic_rule_errors={get(
            errors,
            "config.default_traffic_rule"
          )}
          rules={get(data, "config.rules")}
          routes={get(data, "config.routes")}
          protocol={data.config.protocol}
          onChangeHandler={onChange("config")}
          rules_errors={get(errors, "config.rules")}
        />
      </EuiFlexItem>

      <ResourcesRequirementsPanel
        resources={get(data, "config.resource_request")}
        resourcesOnChangeHandler={onChange("config.resource_request")}
        resourcesErrors={get(errors, "config.resource_request")}
        autoscalingPolicy={get(data, "config.autoscaling_policy")}
        autoscalingPolicyOnChangeHandler={onChange("config.autoscaling_policy")}
        autoscalingPolicyErrors={get(errors, "config.autoscaling_policy")}
        maxAllowedReplica={maxAllowedReplica}
      />
    </EuiFlexGroup>
  );
};

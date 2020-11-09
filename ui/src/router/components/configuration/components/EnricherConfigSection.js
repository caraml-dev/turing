import React from "react";
import { EuiPanel } from "@elastic/eui";
import { DockerConfigViewGroup } from "./docker_config_section/DockerConfigViewGroup";

export const EnricherConfigSection = ({ config: { enricher } }) => {
  return !enricher ? (
    <EuiPanel>Not Configured</EuiPanel>
  ) : (
    <DockerConfigViewGroup componentName="Enricher" dockerConfig={enricher} />
  );
};

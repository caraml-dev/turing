import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { RouterConfigSection } from "./components/RouterConfigSection";
import { ExperimentConfigSection } from "./components/ExperimentConfigSection";
import { EnricherConfigSection } from "./components/EnricherConfigSection";
import { EnsemblerConfigSection } from "./components/EnsemblerConfigSection";
import { LoggingConfigSection } from "./components/LoggingConfigSection";
import { ConfigSection } from "../../../components/config_section";

export const RouterConfigDetails = ({ projectId, config }) => {
  const sections = [
    {
      title: "Router",
      iconType: "bolt",
      children: <RouterConfigSection config={config} />,
    },
    {
      title: "Experiment",
      iconType: "beaker",
      children: (
        <ExperimentConfigSection projectId={projectId} config={config} />
      ),
    },
    {
      title: "Enricher",
      iconType: "package",
      children: <EnricherConfigSection config={config} />,
    },
    {
      title: "Ensembler",
      iconType: "aggregate",
      children: <EnsemblerConfigSection config={config} />,
    },
    {
      title: "Outcome Tracking",
      iconType: "visTagCloud",
      children: <LoggingConfigSection config={config} />,
    },
  ];

  return (
    <EuiFlexGroup direction="column">
      {sections.map((section, idx) => (
        <EuiFlexItem key={`config-section-${idx}`}>
          <ConfigSection title={section.title} iconType={section.iconType}>
            {section.children}
          </ConfigSection>
        </EuiFlexItem>
      ))}
      <EuiSpacer size="l" />
    </EuiFlexGroup>
  );
};

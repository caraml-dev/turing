import React, { useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { ConfigSection } from "../../../components/config_section";
import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { ConfigurationConfigSection } from "./configuration_section/ConfigurationConfigSection";

export const EnsemblingJobConfigView = ({ job }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Jobs",
        href: "../",
      },
      {
        text: `Job ${job.id}`,
        href: "./",
      },
      {
        text: "Details",
      },
    ]);
  }, [job.id]);

  const sections = [
    {
      title: "Source",
      iconType: "importAction",
      children: null,
    },
    {
      title: "Predictions",
      iconType: "heatmap",
      children: null,
    },
    {
      title: "Sink",
      iconType: "exportAction",
      children: null,
    },
    {
      title: "Configuration",
      iconType: "gear",
      children: <ConfigurationConfigSection job={job} />,
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

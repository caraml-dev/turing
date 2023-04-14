import React, { Fragment, useEffect } from "react";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { ConfigSection } from "../../../components/config_section";
import { EuiCallOut, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { ConfigurationConfigSection } from "./configuration_section/ConfigurationConfigSection";
import { SinkConfigSection } from "./sink_config_section/SinkConfigSection";
import { PredictionsConfigSection } from "./prediction_config_section/PredictionsConfigSection";
import { SourceConfigSection } from "./source_config_section/SourceConfigSection";

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
      children: <SourceConfigSection job={job} />,
    },
    {
      title: "Predictions",
      iconType: "heatmap",
      children: <PredictionsConfigSection job={job} />,
    },
    {
      title: "Sink",
      iconType: "exportAction",
      children: <SinkConfigSection job={job} />,
    },
    {
      title: "Configuration",
      iconType: "gear",
      children: <ConfigurationConfigSection job={job} />,
    },
  ];

  return (
    <Fragment>
      {!!job.error && (
        <Fragment>
          <EuiCallOut
            title="Ensembling job has failed"
            color="danger"
            iconType="alert">
            <p>
              <b>Reason: </b>
              {job.error}
            </p>
          </EuiCallOut>

          <EuiSpacer size="m" />
        </Fragment>
      )}
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
    </Fragment>
  );
};

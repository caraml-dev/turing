import React, { Fragment } from "react";
import {
  EuiDescriptionListDescription,
  EuiHorizontalRule,
  EuiLink,
  EuiSpacer,
  EuiText,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import getExperimentUrl from "../../config";
import "./ExperimentsConfigTable.scss";

export const ExperimentsConfigTable = ({ experiments, engineProps }) => {
  return !!experiments.length ? (
    <Fragment>
      {experiments.map((experiment, idx) => {
        const experimentUrl = getExperimentUrl(engineProps, experiment);
        return (
          <Fragment key={experiment.name}>
            {!!experimentUrl ? (
              <EuiTitle size="xxs">
                <EuiTextColor color="secondary">
                  <EuiLink href={experimentUrl} target="_blank" external>
                    <code>{experiment.name}</code>
                  </EuiLink>
                </EuiTextColor>
              </EuiTitle>
            ) : (
              <EuiDescriptionListDescription title={experiment.name}>
                {experiment.name}
              </EuiDescriptionListDescription>
            )}
            {idx < experiments.length - 1 && (
              <Fragment>
                <EuiHorizontalRule size="full" margin="xs" />
                <EuiSpacer size="s" />
              </Fragment>
            )}
          </Fragment>
        );
      })}
    </Fragment>
  ) : (
    <EuiText size="s" color="subdued">
      {engineProps.experiment_selection_enabled ? "None" : "N/A"}
    </EuiText>
  );
};

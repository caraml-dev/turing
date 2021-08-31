import React, { Fragment } from "react";
import { EuiSpacer, EuiSteps } from "@elastic/eui";
import { PredictionSourceConfigSection } from "./PredictionSourceConfigSection";

import "./PredictionsConfigSection.scss";

export const PredictionsConfigSection = ({
  job: {
    job_config: {
      spec: { predictions },
    },
  },
}) => (
  <Fragment>
    <EuiSpacer size="m" />
    <EuiSteps
      className="predictionSources"
      titleSize="xs"
      steps={Object.entries(predictions).map(([id, prediction_source]) => ({
        title: id,
        children: <PredictionSourceConfigSection source={prediction_source} />,
      }))}
    />
  </Fragment>
);

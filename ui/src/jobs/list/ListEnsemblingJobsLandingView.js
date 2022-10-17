import { EuiSpacer, EuiText, EuiPageTemplate } from "@elastic/eui";
import React from "react";
import { useConfig } from "../../config";
import { ListEnsemblingJobsView } from "./ListEnsemblingJobsView";

export const ListEnsemblingJobsLandingView = ({ projectId }) => {
  const {
    appConfig: { batchEnsemblingEnabled },
  } = useConfig();

  return (
    <>
      {batchEnsemblingEnabled ? (
        <ListEnsemblingJobsView projectId={projectId} />
      ) : (
        <EuiPageTemplate>
          <EuiPageTemplate.EmptyPrompt>
            <EuiText>
              Batch ensembling has not been enabled for this deployment of
              Turing.
            </EuiText>
            <EuiSpacer />
            <EuiText>
              To use batch ensembling, please enable it in the configuration
              file used to deploy the Turing app.
            </EuiText>
          </EuiPageTemplate.EmptyPrompt>
        </EuiPageTemplate>
      )}
    </>
  );
};

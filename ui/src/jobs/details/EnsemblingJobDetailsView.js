import { useTuringApi } from "../../hooks/useTuringApi";
import { useInitiallyLoaded } from "../../hooks/useInitiallyLoaded";
import {
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiSpacer,
  EuiPageTemplate,
} from "@elastic/eui";
import React, { Fragment } from "react";
import { PageTitle } from "../../components/page/PageTitle";
import { StatusBadge } from "../../components/status_badge/StatusBadge";
import { JobStatus } from "../../services/job/JobStatus";
import { Redirect, Router } from "@reach/router";
import { EnsemblingJobConfigView } from "./config/EnsemblingJobConfigView";
import { EnsemblingJobDetailsPageHeader } from "../components/job_details_header/EnsemblingJobDetailsPageHeader";
import { EnsemblingJobDetailsPageNavigation } from "../components/page_navigation/EnsemblingJobDetailsPageNavigation";
import { EnsemblingJobLogsView } from "./logs/EnsemblingJobLogsView";

export const EnsemblingJobDetailsView = ({ projectId, jobId, ...props }) => {
  const [{ data: jobDetails, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/jobs/${jobId}`,
    {},
    { config: {} }
  );

  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  return (
    <EuiPageTemplate restrictWidth="90%">
      {!hasInitiallyLoaded ? (
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={true}>
            <EuiLoadingContent lines={3} />
          </EuiFlexItem>
        </EuiFlexGroup>
      ) : error ? (
        <EuiCallOut
          title="Sorry, there was an error"
          color="danger"
          iconType="alert">
          <p>{error.message}</p>
        </EuiCallOut>
      ) : (
        <Fragment>
          <EuiPageTemplate.Header
            bottomBorder={false}
            pageTitle={
              <PageTitle
                title={jobDetails.name}
                icon={"sqlApp"}
                prepend={
                  <StatusBadge
                    status={JobStatus.fromValue(jobDetails.status)}
                  />
                }
              />
            }
          >
            <EnsemblingJobDetailsPageHeader job={jobDetails} {...props} />

            <EuiSpacer size="xs" />

            <EnsemblingJobDetailsPageNavigation job={jobDetails} {...props} />

          </EuiPageTemplate.Header>

          <EuiPageTemplate.Section color={"transparent"}>
            <Router>
              <Redirect from="/" to="details" noThrow />
              <EnsemblingJobConfigView path="details" job={jobDetails} />

              <EnsemblingJobLogsView path="logs" job={jobDetails} />

              <Redirect from="any" to="/error/404" default noThrow />
            </Router>
          </EuiPageTemplate.Section>
        </Fragment>
      )}
    </EuiPageTemplate>
  );
};

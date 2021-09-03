import { useTuringApi } from "../../hooks/useTuringApi";
import { useInitiallyLoaded } from "../../hooks/useInitiallyLoaded";
import {
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
} from "@elastic/eui";
import React, { Fragment } from "react";
import { PageTitle } from "../../components/page/PageTitle";
import { StatusBadge } from "../../components/status_badge/StatusBadge";
import { JobStatus } from "../../services/job/JobStatus";
import { Redirect, Router } from "@reach/router";
import { EnsemblingJobConfigView } from "./config/EnsemblingJobConfigView";
import { EnsemblingJobDetailsPageHeader } from "../components/job_details_header/EnsemblingJobDetailsPageHeader";
import { EnsemblingJobDetailsPageNavigation } from "../components/page_navigation/EnsemblingJobDetailsPageNavigation";

export const EnsemblingJobDetailsView = ({ projectId, jobId, ...props }) => {
  const [{ data: jobDetails, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/jobs/${jobId}`,
    {},
    { config: {} }
  );

  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  return (
    <EuiPage>
      <EuiPageBody>
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
            <EuiPageHeader>
              <EuiPageHeaderSection>
                <PageTitle
                  icon="sqlApp"
                  iconSize="l"
                  size="s"
                  title={jobDetails.name}
                  prepend={
                    <StatusBadge
                      status={JobStatus.fromValue(jobDetails.status)}
                    />
                  }
                />
              </EuiPageHeaderSection>
            </EuiPageHeader>

            <EnsemblingJobDetailsPageHeader job={jobDetails} {...props} />

            <EuiSpacer size="xs" />

            <EnsemblingJobDetailsPageNavigation job={jobDetails} {...props} />

            <EuiSpacer size="m" />

            {!!jobDetails.error && (
              <Fragment>
                <EuiCallOut
                  title="Ensembling job has failed"
                  color="danger"
                  iconType="alert">
                  <p>
                    <b>Reason: </b>
                    {jobDetails.error}
                  </p>
                </EuiCallOut>

                <EuiSpacer size="m" />
              </Fragment>
            )}

            <Router>
              <Redirect from="/" to="details" noThrow />
              <EnsemblingJobConfigView path="details" job={jobDetails} />

              <Redirect from="any" to="/error/404" default noThrow />
            </Router>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

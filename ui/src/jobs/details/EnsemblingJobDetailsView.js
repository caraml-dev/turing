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
import { Navigate, Route, Routes, useParams } from "react-router-dom";
import { PageTitle } from "../../components/page/PageTitle";
import { StatusBadge } from "../../components/status_badge/StatusBadge";
import { JobStatus } from "../../services/job/JobStatus";
import { EnsemblingJobConfigView } from "./config/EnsemblingJobConfigView";
import { EnsemblingJobDetailsPageHeader } from "../components/job_details_header/EnsemblingJobDetailsPageHeader";
import { EnsemblingJobDetailsPageNavigation } from "../components/page_navigation/EnsemblingJobDetailsPageNavigation";
import { EnsemblingJobLogsView } from "./logs/EnsemblingJobLogsView";
import { useConfig } from "../../config";

export const EnsemblingJobDetailsView = () => {
  const { projectId, jobId, "*": section } = useParams();
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [{ data: jobDetails, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/jobs/${jobId}`,
    {},
    { config: {} }
  );

  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
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
            <EnsemblingJobDetailsPageHeader job={jobDetails} />

            <EuiSpacer size="xs" />

            <EnsemblingJobDetailsPageNavigation job={jobDetails} selectedTab={section} />

          </EuiPageTemplate.Header>

          <EuiSpacer size="m" />
          <EuiPageTemplate.Section color={"transparent"}>
            <Routes>
              {/* DETAILS */}
              <Route index element={<Navigate to="details" replace={true} />} />
              <Route path="details" element={<EnsemblingJobConfigView job={jobDetails} />} />
              {/* LOGS */}
              <Route path="logs" element={<EnsemblingJobLogsView job={jobDetails} />} />
            </Routes>
          </EuiPageTemplate.Section>
        </Fragment>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

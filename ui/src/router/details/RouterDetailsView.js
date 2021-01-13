import React, { Fragment, useEffect } from "react";
import {
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer
} from "@elastic/eui";
import { useTuringApi } from "../../hooks/useTuringApi";
import { Redirect, Router } from "@reach/router";
import { RouterConfigView } from "./config/RouterConfigView";
import { RouterDetailsPageHeader } from "./components/router_details_header/RouterDetailsPageHeader";
import { DeploymentStatusBadge } from "../components/status_badge/DeploymentStatusBadge";
import { EditRouterView } from "../edit/EditRouterView";
import { PageTitle } from "../../components/page/PageTitle";
import { RouterDetailsPageNavigation } from "./components/page_navigation/RouterDetailsPageNavigation";
import { RouterActions } from "./components/RouterActions";
import { RouterAlertsView } from "../alerts/RouterAlertsView";
import { Status } from "../../services/status/Status";
import { useInitiallyLoaded } from "../../hooks/useInitiallyLoaded";
import { HistoryView } from "../history/HistoryView";
import { ContainerLogsView } from "../logs/ContainerLogsView";
import { VersionComparisonView } from "../versions/comparison/VersionComparisonView";

export const RouterDetailsView = ({ projectId, routerId, ...props }) => {
  const [{ data: router, isLoaded, error }, fetchRouterDetails] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}`,
    {},
    { config: {} }
  );

  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  useEffect(() => {
    if ((props.location.state || {}).refresh) {
      fetchRouterDetails();
    }
  }, [fetchRouterDetails, props.location.state]);

  useEffect(() => {
    if (router.status === Status.PENDING.toString()) {
      const interval = setInterval(fetchRouterDetails, 5000);
      return () => clearInterval(interval);
    }
  }, [router.status, fetchRouterDetails]);

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
                  title={
                    <Fragment>
                      {router.name}&nbsp;
                      <DeploymentStatusBadge status={router.status} />
                    </Fragment>
                  }
                />
              </EuiPageHeaderSection>
            </EuiPageHeader>

            <RouterDetailsPageHeader router={router} />
            {!(props["*"] === "edit" || props["*"] === "alerts/edit") && (
              <Fragment>
                <EuiSpacer size="xs" />
                <RouterActions
                  onEditRouter={() => props.navigate("./edit")}
                  onDeploySuccess={fetchRouterDetails}
                  onUndeploySuccess={fetchRouterDetails}
                  onDeleteSuccess={() => props.navigate("../")}>
                  {getActions => (
                    <RouterDetailsPageNavigation
                      router={router}
                      actions={getActions(router)}
                      {...props}
                    />
                  )}
                </RouterActions>
              </Fragment>
            )}
            <EuiSpacer size="m" />

            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <RouterConfigView path="details" router={router} />

              <EditRouterView path="edit" router={router} />

              <Redirect from="versions" to="../history" noThrow />
              <HistoryView path="history" router={router} />

              <VersionComparisonView
                path="history/compare/:leftVersionNumber/:rightVersionNumber"
                router={router}
              />

              <RouterAlertsView path="alerts/*" router={router} />

              <ContainerLogsView path="logs" router={router} />

              <Redirect from="any" to="/error/404" default noThrow />
            </Router>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

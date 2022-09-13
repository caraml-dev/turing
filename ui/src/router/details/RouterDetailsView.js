import React, { Fragment, useEffect, useState } from "react";
import {
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiSpacer,
  EuiPageTemplate
} from "@elastic/eui";
import { useTuringApi } from "../../hooks/useTuringApi";
import { Redirect, Router } from "@reach/router";
import { RouterConfigView } from "./config/RouterConfigView";
import { RouterDetailsPageHeader } from "./components/router_details_header/RouterDetailsPageHeader";
import { StatusBadge } from "../../components/status_badge/StatusBadge";
import { EditRouterView } from "../edit/EditRouterView";
import { PageTitle } from "../../components/page/PageTitle";
import { RouterDetailsPageNavigation } from "./components/page_navigation/RouterDetailsPageNavigation";
import { RouterActions } from "./components/RouterActions";
import { RouterAlertsView } from "../alerts/RouterAlertsView";
import { Status } from "../../services/status/Status";
import { useInitiallyLoaded } from "../../hooks/useInitiallyLoaded";
import { HistoryView } from "../history/HistoryView";
import { RouterLogsView } from "./logs/RouterLogsView";
import { VersionComparisonView } from "../versions/comparison/VersionComparisonView";
import { TuringRouter } from "../../services/router/TuringRouter";
import { useConfig } from "../../config";

export const RouterDetailsView = ({ projectId, routerId, ...props }) => {
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [router, setRouter] = useState({});
  const [{ data: routerDetails, isLoaded, error }, fetchRouterDetails] =
    useTuringApi(
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
    if (router.status?.toString() === Status.PENDING.toString()) {
      const interval = setInterval(fetchRouterDetails, 5000);
      return () => clearInterval(interval);
    }
  }, [router.status, fetchRouterDetails]);

  useEffect(() => {
    // Parse router details
    if (!!routerDetails) {
      setRouter(TuringRouter.fromJson(routerDetails));
    }
  }, [routerDetails, setRouter]);

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
                title={router.name}
                prepend={<StatusBadge status={router.status} />}
              />
            }
          >
            <RouterDetailsPageHeader router={router} />

            {!(props["*"] === "edit" || props["*"] === "alerts/edit") && (
              <Fragment>
                <EuiSpacer size="xs" />
                <RouterActions
                  onEditRouter={() => props.navigate("./edit")}
                  onDeploySuccess={fetchRouterDetails}
                  onUndeploySuccess={fetchRouterDetails}
                  onDeleteSuccess={() => props.navigate("../")}>
                  {(getActions) => (
                    <RouterDetailsPageNavigation
                      router={router}
                      actions={getActions(router)}
                      {...props}
                    />
                  )}
                </RouterActions>
              </Fragment>
            )}
          </EuiPageTemplate.Header>

          <EuiSpacer size="m" />
          <EuiPageTemplate.Section color={"transparent"}>
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

              <RouterLogsView path="logs" router={router} />

              <Redirect from="any" to="/error/404" default noThrow />
            </Router>
          </EuiPageTemplate.Section>
        </Fragment>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

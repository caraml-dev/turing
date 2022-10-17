import React, { Fragment, useEffect, useState } from "react";
import {
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiSpacer,
  EuiPageTemplate
} from "@elastic/eui";
import { Navigate, Route, Routes, useLocation, useNavigate, useParams } from "react-router-dom";
import { useTuringApi } from "../../hooks/useTuringApi";
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

export const RouterDetailsView = () => {
  const { projectId, routerId, "*": section } = useParams();
  const location = useLocation();
  const navigate = useNavigate();

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
    if ((location.state || {}).refresh) {
      fetchRouterDetails();
    }
  }, [fetchRouterDetails, location.state]);

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

            {!(section === "edit" || section === "alerts/edit") && (
              <Fragment>
                <EuiSpacer size="xs" />
                <RouterActions
                  onEditRouter={() => navigate("./edit")}
                  onDeploySuccess={fetchRouterDetails}
                  onUndeploySuccess={fetchRouterDetails}
                  onDeleteSuccess={() => navigate("../")}>
                  {(getActions) => (
                    <RouterDetailsPageNavigation
                      router={router}
                      actions={getActions(router)}
                      selectedTab={section}
                    />
                  )}
                </RouterActions>
              </Fragment>
            )}
          </EuiPageTemplate.Header>

          <EuiSpacer size="m" />
          <EuiPageTemplate.Section color={"transparent"}>
            <Routes>
              {/* DETAILS */}
              <Route index element={<Navigate to="details" replace={true} />} />
              <Route path="details" element={<RouterConfigView router={router} />} />
              {/* EDIT */}
              <Route path="edit" element={<EditRouterView projectId={projectId} router={router} />} />
              {/* HISTORY */}
              <Route path="versions" element={<Navigate to="../history" replace={true} />} />
              <Route path="history" element={<HistoryView projectId={projectId} router={router} />} />
              {/* DIFF */}
              <Route path="history/compare/:leftVersionNumber/:rightVersionNumber" element={<VersionComparisonView router={router} />} />
              {/* ALERTS */}
              <Route path="alerts/*" element={<RouterAlertsView projectId={projectId} router={router} />} />
              {/* LOGS */}
              <Route path="logs" element={<RouterLogsView path="logs" router={router} />} />
            </Routes>
          </EuiPageTemplate.Section>
        </Fragment>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

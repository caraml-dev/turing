import React, { Fragment, useCallback, useEffect, useState } from "react";
import {
  EuiBadge,
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiSpacer,
  EuiPageTemplate
} from "@elastic/eui";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { Redirect, Router } from "@reach/router";
import { RouterVersionConfigView } from "./config/RouterVersionConfigView";
import { RouterVersionDetailsPageNavigation } from "./page_navigation/RouterVersionDetailsPageNavigation";
import { PageTitle } from "../../../components/page/PageTitle";
import { VersionDetailsPageHeader } from "./version_details_header/VersionDetailsPageHeader";
import { RouterVersionActions } from "../components/RouterVersionActions";
import { Status } from "../../../services/status/Status";
import { useInitiallyLoaded } from "../../../hooks/useInitiallyLoaded";

import { TuringRouter } from "../../../services/router/TuringRouter";
import { RouterVersion } from "../../../services/version/RouterVersion";

export const RouterVersionDetailsView = ({
  projectId,
  routerId,
  versionId,
  ...props
}) => {
  // Use local states to store parsed responses
  const [router, setRouter] = useState({});
  const [version, setVersion] = useState({});

  const [{ data: versionDetails, isLoaded, error }, fetchVersionDetails] =
    useTuringApi(
      `/projects/${projectId}/routers/${routerId}/versions/${versionId}`,
      {},
      { config: {} }
    );
  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  const [{ data: routerDetails }, fetchRouterDetails] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}`,
    {},
    {}
  );

  const refreshData = useCallback(() => {
    fetchRouterDetails();
    fetchVersionDetails();
  }, [fetchRouterDetails, fetchVersionDetails]);

  const isActiveConfig = version.version === (router.config || {}).version;

  useEffect(() => {
    if (router.status?.toString() === Status.PENDING.toString()) {
      const interval = setInterval(refreshData, 5000);
      return () => clearInterval(interval);
    }
  }, [router.status, refreshData]);

  useEffect(() => {
    if (!!routerDetails) {
      setRouter(TuringRouter.fromJson(routerDetails));
    }
  }, [routerDetails, setRouter]);

  useEffect(() => {
    if (!!versionDetails) {
      setVersion(RouterVersion.fromJson(versionDetails));
    }
  }, [versionDetails, setVersion]);

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize={"none"}>
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
                icon="graphApp"
                title={
                  <Fragment>
                    Version <b>{version.version}</b>
                    &nbsp;&nbsp;
                    {isActiveConfig && (
                      <EuiBadge
                        color="default"
                        style={{ letterSpacing: "initial" }}>
                        Current
                      </EuiBadge>
                    )}
                  </Fragment>
                }
              />
            }
          >
            <VersionDetailsPageHeader version={version} />

            <EuiSpacer size="xs" />

            <RouterVersionActions
              router={router}
              onDeploySuccess={refreshData}
              onDeleteSuccess={() => props.navigate("../")}>
              {(actions) => (
                <RouterVersionDetailsPageNavigation
                  version={version}
                  actions={actions.map((action) => ({
                    ...action,
                    onClick: () => action.onClick(version),
                    hidden: !action.available(version),
                    disabled: !action.enabled(version),
                  }))}
                  {...props}
                />
              )}
            </RouterVersionActions>
          </EuiPageTemplate.Header>

          <EuiSpacer size="m" />
          <EuiPageTemplate.Section color={"transparent"}>
            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <RouterVersionConfigView path="details" config={version} />
              <Redirect from="any" to="/error/404" default noThrow />
            </Router>
          </EuiPageTemplate.Section>
        </Fragment>
      )}
    </EuiPageTemplate>
  );
};

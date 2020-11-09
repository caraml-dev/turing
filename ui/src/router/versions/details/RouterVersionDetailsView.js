import React, { Fragment, useCallback, useEffect } from "react";
import {
  EuiBadge,
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
import { useTuringApi } from "../../../hooks/useTuringApi";
import { Redirect, Router } from "@reach/router";
import { RouterVersionConfigView } from "./config/RouterVersionConfigView";
import { RouterVersionDetailsPageNavigation } from "./page_navigation/RouterVersionDetailsPageNavigation";
import { PageTitle } from "../../../components/page/PageTitle";
import { VersionDetailsPageHeader } from "./version_details_header/VersionDetailsPageHeader";
import { RouterVersionActions } from "../components/RouterVersionActions";
import { Status } from "../../../services/status/Status";
import { useInitiallyLoaded } from "../../../hooks/useInitiallyLoaded";

export const RouterVersionDetailsView = ({
  projectId,
  routerId,
  versionId,
  ...props
}) => {
  const [
    { data: version, isLoaded, error },
    fetchVersionDetails
  ] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/versions/${versionId}`,
    {},
    { config: {} }
  );
  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  const [{ data: router }, fetchRouterDetails] = useTuringApi(
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
    if (router.status === Status.PENDING.toString()) {
      const interval = setInterval(refreshData, 5000);
      return () => clearInterval(interval);
    }
  }, [router.status, refreshData]);

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
                  size="m"
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
              </EuiPageHeaderSection>
            </EuiPageHeader>

            <VersionDetailsPageHeader version={version} />

            <EuiSpacer size="xs" />

            <RouterVersionActions
              router={router}
              onDeploySuccess={refreshData}
              onDeleteSuccess={() => props.navigate("../")}>
              {actions => (
                <RouterVersionDetailsPageNavigation
                  version={version}
                  actions={actions.map(action => ({
                    ...action,
                    onClick: () => action.onClick(version),
                    hidden: !action.available(version),
                    disabled: !action.enabled(version)
                  }))}
                  {...props}
                />
              )}
            </RouterVersionActions>
            <EuiSpacer size="m" />

            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <RouterVersionConfigView path="details" config={version} />
              <Redirect from="any" to="/error/404" default noThrow />
            </Router>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

import React, { useEffect, useState } from "react";
import { ConfigSection } from "../../components/configuration/components/section";
import {
  EuiCallOut,
  EuiFilterButton,
  EuiFilterGroup,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingChart,
  EuiPanel,
  EuiSpacer,
  EuiTextAlign
} from "@elastic/eui";
import ReactDiffViewer, { DiffMethod } from "react-diff-viewer";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { RouterVersion } from "../../../services/version/RouterVersion";

export const VersionComparisonView = ({
  router,
  leftVersionNumber,
  rightVersionNumber,
  location: { state },
  ...props
}) => {
  const { versionLeft, versionRight } = state || {};

  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState(undefined);
  const [splitView, setSplitView] = useState(true);

  const [leftVersion] = useTuringApi(
    `/projects/${props.projectId}/routers/${props.routerId}/versions/${leftVersionNumber}`,
    {},
    versionLeft,
    !versionLeft
  );

  const [rightVersion] = useTuringApi(
    `/projects/${props.projectId}/routers/${props.routerId}/versions/${rightVersionNumber}`,
    {},
    versionRight,
    !versionRight
  );

  useEffect(() => {
    if (!!leftVersion.data && !!rightVersion.data) {
      setIsLoaded(true);
    } else if (!!leftVersion.error || !!rightVersion.error) {
      setIsLoaded(true);
      setError(leftVersion.error || rightVersion.error);
    }
  }, [leftVersion, rightVersion]);

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: `../../../../`
      },
      {
        text: router.name,
        href: `../../../`
      },
      {
        text: "Versions",
        href: `../../`
      },
      {
        text: "Compare"
      }
    ]);
  }, [router.name]);

  return (
    <ConfigSection title="Comparing Versions">
      <EuiPanel>
        <EuiFlexGroup direction="row" justifyContent="flexEnd">
          <EuiFlexItem grow={false}>
            <EuiFilterGroup>
              <EuiFilterButton
                hasActiveFilters={!splitView}
                onClick={() => setSplitView(false)}>
                Unified
              </EuiFilterButton>
              <EuiFilterButton
                hasActiveFilters={splitView}
                onClick={() => setSplitView(true)}>
                Split
              </EuiFilterButton>
            </EuiFilterGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="s" />

        {!isLoaded ? (
          <EuiTextAlign textAlign="center">
            <EuiLoadingChart size="xl" mono />
          </EuiTextAlign>
        ) : error ? (
          <EuiCallOut
            title="Sorry, there was an error"
            color="danger"
            iconType="alert">
            <p>{error.message}</p>
          </EuiCallOut>
        ) : (
          <ReactDiffViewer
            leftTitle={`Version ${leftVersionNumber}`}
            rightTitle={`Version ${rightVersionNumber}`}
            oldValue={RouterVersion.fromJson(leftVersion.data).toPrettyYaml()}
            newValue={RouterVersion.fromJson(rightVersion.data).toPrettyYaml()}
            styles={{
              line: {
                wordBreak: "break-word",
                fontSize: "0.775rem"
              }
            }}
            compareMethod={DiffMethod.WORDS_WITH_SPACE}
            splitView={splitView}
          />
        )}
      </EuiPanel>
    </ConfigSection>
  );
};

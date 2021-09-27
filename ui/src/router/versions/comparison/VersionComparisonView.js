import React, { useEffect, useState } from "react";
import { ConfigSection } from "../../../components/config_section";
import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPanel,
  EuiTextAlign,
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { RouterVersion } from "../../../services/version/RouterVersion";
import { VersionComparisonPanel } from "../components/version_diff/VersionComparisonPanel";

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
        href: `../../../../`,
      },
      {
        text: router.name,
        href: `../../../`,
      },
      {
        text: "Versions",
        href: `../../`,
      },
      {
        text: "Compare",
      },
    ]);
  }, [router.name]);

  return (
    <ConfigSection title="Comparing Versions">
      <EuiPanel>
        {!isLoaded ? (
          <EuiTextAlign textAlign="center">
            <EuiLoadingChart size="xl" mono />
          </EuiTextAlign>
        ) : error ? (
          <EuiCallOut
            title="Sorry, there was an error"
            color="danger"
            iconType="alert"
          >
            <p>{error.message}</p>
          </EuiCallOut>
        ) : (
          <VersionComparisonPanel
            leftValue={RouterVersion.fromJson(leftVersion.data).toPrettyYaml()}
            rightValue={RouterVersion.fromJson(
              rightVersion.data
            ).toPrettyYaml()}
            leftTitle={`Version ${leftVersionNumber}`}
            rightTitle={`Version ${rightVersionNumber}`}
          />
        )}
      </EuiPanel>
    </ConfigSection>
  );
};

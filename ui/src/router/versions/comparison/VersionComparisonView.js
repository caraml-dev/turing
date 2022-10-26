import React, { useContext, useEffect, useState } from "react";
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
import { ExperimentEngineContextProvider } from "../../../providers/experiments/ExperimentEngineContextProvider";
import ExperimentEngineContext from "../../../providers/experiments/context";
import { useLocation, useParams } from "react-router-dom";

const VersionComparisonView = ({ router }) => {
  const { projectId, routerId, leftVersionNumber, rightVersionNumber } = useParams();
  const location = useLocation();

  const { versionLeft, versionRight } = location.state || {};

  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState(undefined);

  const [leftVersion] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/versions/${leftVersionNumber}`,
    {},
    versionLeft,
    !versionLeft
  );

  const [rightVersion] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/versions/${rightVersionNumber}`,
    {},
    versionRight,
    !versionRight
  );

  const { getEngineProperties, isLoaded: isExpCtxLoaded } = useContext(
    ExperimentEngineContext
  );

  useEffect(() => {
    if (!!leftVersion.data && !!rightVersion.data && isExpCtxLoaded) {
      setIsLoaded(true);
    } else if (!!leftVersion.error || !!rightVersion.error) {
      setIsLoaded(true);
      setError(leftVersion.error || rightVersion.error);
    }
  }, [leftVersion, rightVersion, isExpCtxLoaded]);

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
            iconType="alert">
            <p>{error.message}</p>
          </EuiCallOut>
        ) : (
          <VersionComparisonPanel
            leftValue={RouterVersion.fromJson(leftVersion.data).toPrettyYaml({
              experiment_engine: getEngineProperties(
                leftVersion.data?.experiment_engine?.type
              ),
            })}
            rightValue={RouterVersion.fromJson(rightVersion.data).toPrettyYaml({
              experiment_engine: getEngineProperties(
                rightVersion.data?.experiment_engine?.type
              ),
            })}
            leftTitle={`Version ${leftVersionNumber}`}
            rightTitle={`Version ${rightVersionNumber}`}
          />
        )}
      </EuiPanel>
    </ConfigSection>
  );
};

const VersionComparisonViewWrapper = (props) => (
  <ExperimentEngineContextProvider>
    <VersionComparisonView {...props} />
  </ExperimentEngineContextProvider>
);

export { VersionComparisonViewWrapper as VersionComparisonView };

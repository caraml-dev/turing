import React, { useEffect, useRef } from "react";
import { EuiEmptyPrompt, EuiLoadingChart } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { RouterConfigDetails } from "../../components/configuration/RouterConfigDetails";
import { ExperimentEngineContextProvider } from "../../../providers/experiments/ExperimentEngineContextProvider";
import { OverlayMask } from "../../../components/overlay_mask/OverlayMask";
import { Status } from "../../../services/status/Status";

export const RouterConfigView = ({ router }) => {
  const configSectionRef = useRef();

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: "../",
      },
      {
        text: router.name,
        href: "./",
      },
      {
        text: "Configuration",
      },
    ]);
  }, [router.name]);

  return router.config ? (
    <div ref={configSectionRef}>
      {[Status.UNDEPLOYED, Status.PENDING].includes(router.status) && (
        <OverlayMask parentRef={configSectionRef} opacity={0.4}>
          {router.status === Status.PENDING && (
            <EuiLoadingChart size="xl" mono />
          )}
        </OverlayMask>
      )}
      <ExperimentEngineContextProvider>
        <RouterConfigDetails
          projectId={router.project_id}
          config={router.config}
        />
      </ExperimentEngineContextProvider>
    </div>
  ) : (
    <EuiEmptyPrompt
      title={<h2>Router is not deployed</h2>}
      body={
        <p>
          Deploy it first and wait for the deployment to complete before you can
          see the configuration details here
        </p>
      }
    />
  );
};

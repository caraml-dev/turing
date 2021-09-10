import React, { useEffect, useMemo, useRef } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { RouterConfigDetails } from "../../components/configuration/RouterConfigDetails";
import { EuiEmptyPrompt, EuiLoadingChart } from "@elastic/eui";
import { OverlayMask } from "../../../components/overlay_mask/OverlayMask";
import { Status } from "../../../services/status/Status";

export const RouterConfigView = ({ router }) => {
  const configSectionRef = useRef();

  const status = useMemo(
    () => Status.fromValue(router.status),
    [router.status]
  );

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
      {[Status.UNDEPLOYED, Status.PENDING].includes(status) && (
        <OverlayMask parentRef={configSectionRef} opacity={0.4}>
          {status === Status.PENDING && <EuiLoadingChart size="xl" mono />}
        </OverlayMask>
      )}
      <RouterConfigDetails config={router.config} />
    </div>
  ) : (
    <EuiEmptyPrompt
      title={<h2>Router is not deployed</h2>}
      body={
        <p>
          Deploy it first and wait for deployment to complete before you can see
          the configuration details here
        </p>
      }
    />
  );
};

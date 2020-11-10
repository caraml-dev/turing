import React, { Fragment, useMemo, useRef } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiButton,
  EuiText,
  EuiSpacer,
  EuiLoadingChart
} from "@elastic/eui";
import { AlertConfigSection } from "../components/alert_config_section/AlertConfigSection";
import {
  ConfigSectionPanel,
  ConfigSectionTitle
} from "../../components/configuration/components/section";
import { Status } from "../../../services/status/Status";
import { OverlayMask } from "../../../components/overlay_mask/OverlayMask";
import { supportedAlerts } from "../config";

export const RouterAlertDetails = ({ alertsData, routerStatus, ...props }) => {
  const alertDetailsRef = useRef();
  const status = useMemo(() => Status.fromValue(routerStatus), [routerStatus]);

  return (
    <div ref={alertDetailsRef}>
      {[Status.UNDEPLOYED, Status.PENDING].includes(status) && (
        <OverlayMask parentRef={alertDetailsRef} opacity={0.4}>
          {status === Status.PENDING && <EuiLoadingChart size="xl" mono />}
        </OverlayMask>
      )}
      <EuiFlexGroup alignItems="baseline">
        <EuiFlexItem>
          <ConfigSectionTitle
            title={
              <Fragment>
                Alerts for team <b>{alertsData.team}</b>
              </Fragment>
            }
          />
        </EuiFlexItem>
        <EuiFlexItem grow={false}>
          <EuiButton
            size="s"
            onClick={() => props.navigate("./edit")}
            disabled={status !== Status.DEPLOYED}>
            Configure Alerts
          </EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
      <EuiSpacer size="l" />

      <EuiFlexGroup>
        {supportedAlerts.map((alertType, idx) => (
          <EuiFlexItem key={`config-section-${idx}`}>
            <ConfigSectionPanel title={alertType.title}>
              {!!alertsData.alerts[alertType.metric] ? (
                <AlertConfigSection
                  alert={alertsData.alerts[alertType.metric]}
                  unit={alertType.unit}
                />
              ) : (
                <EuiText>Not Configured</EuiText>
              )}
            </ConfigSectionPanel>
          </EuiFlexItem>
        ))}
      </EuiFlexGroup>
    </div>
  );
};

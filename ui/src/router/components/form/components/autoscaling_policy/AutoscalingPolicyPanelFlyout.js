import React from "react";
import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFlyout,
  EuiFlyoutBody,
  EuiFlyoutFooter,
  EuiFlyoutHeader,
  EuiText,
  EuiTitle,
} from "@elastic/eui";

export const AutoscalingPolicyPanelFlyout = ({ onClose }) => {
  return (
    <EuiFlyout
      onClose={() => onClose()}
      aria-labelledby="autoscaling-policy-help"
      size="s"
      ownFocus>
      <EuiFlyoutHeader hasBorder>
        <EuiTitle size="s">
          <h2 id="autoscaling-policy-help">Autoscaling Policy</h2>
        </EuiTitle>
      </EuiFlyoutHeader>

      <EuiFlyoutBody>
        <EuiText size="s">
          <p>
            The Autoscaling Policy defines the condition for increasing or decreasing the number of
            replicas, for users that require granular control over the behavior. Currently, 4 metrics
            are supported: Concurrency, RPS, CPU and Memory. <b>Concurrency</b> is the default metric,
            with a target value of <b>1</b>.
          </p>

          <h3>Guidelines for tuning the Autoscaling Target</h3>

          <p>
            For most users, the default values should suffice.
            For more advanced users, the following guidelines may be useful in tuning the autoscaling parameters:
          </p>
          <ul>
            <li>
              <b>RPS:</b> Suitable if a single replica has been tested to perform well up to a given RPS.
              Users can deploy a single replica with a reasonable CPU/Memory configuration, slowly increase the
              request rate to the deployment to determine the boundary at which the performance begins to dip.
            </li>
            <li>
              <b>Concurrency:</b> A process similar to RPS described above may be followed to tune the target value.
            </li>
            <li>
              <b>CPU / Memory:</b> These can be chosen according to whether the process is CPU bound
              or memory bound. With this, scale to 0 is not supported.
            </li>
          </ul>
        </EuiText>
      </EuiFlyoutBody>

      <EuiFlyoutFooter>
        <EuiFlexGroup justifyContent="spaceBetween">
          <EuiFlexItem grow={false}>
            <EuiButtonEmpty
              iconType="cross"
              onClick={() => onClose()}
              flush="left">
              Close
            </EuiButtonEmpty>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlyoutFooter>
    </EuiFlyout>
  );
};

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

export const RulesPanelFlyout = ({ onClose }) => {
  return (
    <EuiFlyout
      onClose={() => onClose()}
      aria-labelledby="traffic-rules-help"
      size="s"
      ownFocus
    >
      <EuiFlyoutHeader hasBorder>
        <EuiTitle size="s">
          <h2 id="traffic-rules-help">Traffic Rules</h2>
        </EuiTitle>
      </EuiFlyoutHeader>

      <EuiFlyoutBody>
        <EuiText size="s">
          <p>
            Traffic rules define which routes should be "activated" for a
            particular request to your router. Each rule is defined by one or
            more request conditions, and one or more routes that would be
            activated if the request satisfies all conditions in the rule.
          </p>
          <p>Each route may be associated with zero or more traffic rules.</p>
          <ul>
            <li>
              If the route is attached to a single traffic rule, then Turing
              will only send a request to this route if the request meets this
              rule's conditions.
            </li>
            <li>
              If the route is not attached to any of the traffic rules, then
              Turing will call this route with every incoming request.
            </li>
            <li>
              If the route is attached to multiple rules and the request
              satisfies more than one rule, then Turing will decide what group
              of routes should receive this request based on the order in which
              the traffic rules are defined.
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
              flush="left"
            >
              Close
            </EuiButtonEmpty>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlyoutFooter>
    </EuiFlyout>
  );
};

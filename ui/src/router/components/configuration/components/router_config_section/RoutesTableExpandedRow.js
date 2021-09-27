import React from "react";
import {
  EuiAvatar,
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHorizontalRule,
  EuiPanel,
} from "@elastic/eui";
import { TrafficRuleCondition } from "../../../traffic_rule_condition/TrafficRuleCondition";

const OrDivider = () => (
  <EuiFlexGroup direction="row" gutterSize="s" alignItems="center">
    <EuiFlexItem grow={1}>
      <EuiHorizontalRule size="full" />
    </EuiFlexItem>

    <EuiFlexItem grow={false}>
      <EuiAvatar size="m" color="#F5F7FA" name="O R" />
    </EuiFlexItem>

    <EuiFlexItem grow={1}>
      <EuiHorizontalRule size="full" />
    </EuiFlexItem>
  </EuiFlexGroup>
);

export const RoutesTableExpandedRow = ({ item }) => {
  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      <EuiFlexItem>
        <EuiCallOut
          color="warning"
          title="Turing will only send a request to this route if the request meets
            all the conditions from at least one of the rules shown below."
          iconType="questionInCircle"
        />
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiFlexGroup direction="row" justifyContent="center" wrap>
          <EuiFlexItem style={{ maxWidth: "80%" }}>
            <EuiFlexGroup direction="column" gutterSize="none">
              {item.rules.map((rule, idx) => (
                <EuiFlexItem key={`rule-${idx}`}>
                  <EuiPanel>
                    <EuiFlexGroup direction="column" gutterSize="s">
                      {rule.map((condition, idx) => (
                        <EuiFlexItem key={`rule-condition-${idx}`}>
                          <TrafficRuleCondition
                            condition={condition}
                            readOnly={true}
                          />
                        </EuiFlexItem>
                      ))}
                    </EuiFlexGroup>
                  </EuiPanel>

                  {idx < item.rules.length - 1 && <OrDivider />}
                </EuiFlexItem>
              ))}
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
      <EuiFlexItem />
    </EuiFlexGroup>
  );
};

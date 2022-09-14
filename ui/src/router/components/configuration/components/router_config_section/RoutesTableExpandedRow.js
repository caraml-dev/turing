import React from "react";
import {
  EuiAvatar,
  EuiCallOut,
  EuiBadge,
  EuiCard,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFlexGrid,
  EuiHorizontalRule,
  EuiSpacer,
  EuiText,
} from "@elastic/eui";

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

export const RoutesTableExpandedRow = ({ item, inDefaultTrafficRule }) => {
  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      <EuiFlexItem>
        {item.rules && <EuiCallOut
          color="warning"
          title="Turing will only send a request to this route if the request meets
            all the conditions from at least one of the rules shown below."
          iconType="questionInCircle"
        />}
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiFlexGroup direction="row" justifyContent="center" wrap>
          <EuiFlexItem style={{ maxWidth: "70%" }}>
            <EuiFlexGroup direction="column" gutterSize="none">
              {item.rules && item.rules.map((rule, idx) => (
                <EuiFlexItem key={`rule-${idx}`}>
                  <EuiCard title={rule.name} layout="horizontal">
                    <EuiHorizontalRule margin="xs" />
                    <EuiFlexGroup direction="column" gutterSize="s">
                      {rule.conditions.map((condition, idx) => (
                        <EuiFlexItem key={`rule-condition-${idx}`}>
                          <EuiFlexGroup direction="row" gutterSize="m" alignItems="baseline">
                            <EuiFlexItem grow={1}>
                              <EuiFlexGroup direction="row" gutterSize="m" alignItems="baseline">
                              <EuiFlexItem grow={false}>
                                <EuiBadge color="primary">
                                  {condition.field_source}
                                </EuiBadge>
                              </EuiFlexItem>
                              <EuiFlexItem grow={false}>
                                <EuiBadge>
                                  {condition.field}
                                </EuiBadge>
                              </EuiFlexItem>
                              </EuiFlexGroup>
                            </EuiFlexItem>

                            <EuiFlexItem grow={false}>
                              <EuiBadge color="warning">
                                {condition.operator}
                              </EuiBadge>
                            </EuiFlexItem>
                              
                            <EuiFlexItem grow={1}>
                              <EuiFlexGrid columns={4} gutterSize="s">
                                {condition.values.map((val, idx) => (
                                  <EuiFlexItem key={`filter-${idx}`} grow={false}>
                                    <EuiBadge>
                                      {val}
                                    </EuiBadge>
                                  </EuiFlexItem>
                                ))}
                                </EuiFlexGrid>
                            </EuiFlexItem>
                          </EuiFlexGroup>
                            <EuiSpacer size="xs" />
                        </EuiFlexItem>
                      ))}
                    </EuiFlexGroup>
                  </EuiCard>

                  {idx < item.rules.length - 1 && <OrDivider />}
                </EuiFlexItem>
              ))}
              {inDefaultTrafficRule && (
                <>
                {item.rules && <OrDivider />}
                <EuiFlexItem key="default-rule">
                  <EuiCard
                    title="default-traffic-rule"
                    titleSize="xs"
                    description={
                      <EuiText>
                        This route is {item.rules && "also"} a part of the Default Rule. This route will receive the request if none of the routing rules' conditions are met.
                      </EuiText>
                    }
                    layout="horizontal"
                    display="warning"
                  />
                </EuiFlexItem>
                </>
              )}
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
      <EuiFlexItem />
    </EuiFlexGroup>
  );
};

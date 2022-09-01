import React, { Fragment, useEffect, useState } from "react";
import { Panel } from "../Panel";
import {
  EuiButton,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiLink,
  EuiSpacer,
  EuiToolTip,
} from "@elastic/eui";
import { get } from "../../../../../components/form/utils";
import { RulesPanelFlyout } from "./RulesPanelFlyout";
import { RuleCard } from "./rule_card/RuleCard";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { newRule } from "../../../../../services/router/TuringRouter";

const defaultRuleName = "default";

export const RulesPanel = ({ rules, routes, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddRule = () => {
    // Default rule should always be last
    const orderedRules = rules.filter(rule => rule.name !== defaultRuleName);
    const defaultRule = rules.filter(rule => rule.name === defaultRuleName);

    onChange("rules")([...orderedRules, newRule(), ...defaultRule]);
  };

  const onDeleteRule = (idx) => () => {
    rules.splice(idx, 1);
    
    // If no custom rule left, remove default rule
    if (rules.length === 1 && rules.at(0).name === defaultRuleName) {
      rules = [];
    }
    onChange("rules")(rules);
  };

  const addRuleButton = (
    <EuiButton fullWidth onClick={onAddRule} isDisabled={routes.length < 2}>
      + Add Rule
    </EuiButton>
  );

  const [isFlyoutVisible, setIsFlyoutVisible] = useState(false);

  useEffect(() => {
    // If there are custom rules and no default rule, add default rule
    if (rules.length > 0 && rules.at(-1).name !== defaultRuleName) {
      const defaultRule = newRule();
      defaultRule.name = defaultRuleName;
      onChange("rules")([...rules, defaultRule]);
    }
  }, [rules, routes, onChange])

  return (
    <Panel
      title={
        <Fragment>
          Traffic Rules{" "}
          <EuiLink color="ghost" onClick={() => setIsFlyoutVisible(true)}>
            <EuiIcon
              type="questionInCircle"
              color="subdued"
              className="eui-alignBaseline"
            />
          </EuiLink>
        </Fragment>
      }>
      <EuiFlexGroup direction="column" gutterSize="s">
        {rules.map((rule, idx) => (
          <EuiFlexItem key={`rule-${idx}`}>
            <RuleCard
              isDefault={idx === Object.keys(rules).length - 1}
              rule={rule}
              routes={routes}
              onChangeHandler={onChange(`rules.${idx}`)}
              onDelete={onDeleteRule(idx)}
              errors={get(errors, `${idx}`)}
            />
            <EuiSpacer size="s" />
          </EuiFlexItem>
        ))}
        <EuiFlexItem>
          {routes.length < 2 ? (
            <EuiToolTip content="You should have more than one route in order to be able to define traffic rules">
              {addRuleButton}
            </EuiToolTip>
          ) : (
            addRuleButton
          )}
        </EuiFlexItem>
      </EuiFlexGroup>

      {isFlyoutVisible && (
        <RulesPanelFlyout onClose={() => setIsFlyoutVisible(false)} />
      )}
    </Panel>
  );
};

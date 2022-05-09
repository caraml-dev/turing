import React, { Fragment, useState } from "react";
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

export const RulesPanel = ({ rules, routes, onChangeHandler, errors = {} }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddRule = () => {
    onChange("rules")([...rules, newRule()]);
  };

  const onDeleteRule = (idx) => () => {
    rules.splice(idx, 1);
    onChange("rules")(rules);
  };

  const addRuleButton = (
    <EuiButton fullWidth onClick={onAddRule} isDisabled={!routes.length}>
      + Add Rule
    </EuiButton>
  );

  const [isFlyoutVisible, setIsFlyoutVisible] = useState(false);

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
            <EuiToolTip content="You should have other routes besides the default one in order to be able to define traffic rules">
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

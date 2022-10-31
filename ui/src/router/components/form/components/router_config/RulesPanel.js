import React, { Fragment } from "react";
import { useToggle } from "@gojek/mlp-ui";
import { Panel } from "../Panel";
import {
  EuiButton,
  EuiDragDropContext,
  euiDragDropReorder,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHorizontalRule,
  EuiIcon,
  EuiLink,
  EuiSpacer,
  EuiToolTip,
} from "@elastic/eui";
import { get } from "../../../../../components/form/utils";
import { RulesPanelFlyout } from "./RulesPanelFlyout";
import { RuleCard } from "./rule_card/RuleCard";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import {
  newDefaultRule,
  newRule,
} from "../../../../../services/router/TuringRouter";

export const RulesPanel = ({
  default_traffic_rule,
  rules,
  routes,
  protocol,
  onChangeHandler,
  rules_errors = {},
  default_traffic_rule_errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddRule = () => {
    onChange("rules")([...rules, newRule()]);
    // Default rule should only be added when there are custom rules
    if (!default_traffic_rule) {
      onChange("default_traffic_rule")(newDefaultRule());
    }
  };

  const onDeleteRule = (idx) => () => {
    rules.splice(idx, 1);

    // If no custom rule left, remove default rule
    if (rules.length === 0) {
      onChange("default_traffic_rule")(null);
    }
    onChange("rules")(rules);
  };

  const addRuleButton = (
    <EuiButton fullWidth onClick={onAddRule} isDisabled={routes.length < 2}>
      + Add Rule
    </EuiButton>
  );

  const [isFlyoutVisible, toggleIsFlyoutVisible] = useToggle();

  const onDragEnd = ({ source, destination }) => {
    if (source && destination) {
      const items = euiDragDropReorder(rules, source.index, destination.index);
      onChange("rules")(items);
    }
  };

  return (
    <Panel
      title={
        <Fragment>
          Traffic Rules{" "}
          <EuiLink color="ghost" onClick={() => toggleIsFlyoutVisible()}>
            <EuiIcon
              type="questionInCircle"
              color="subdued"
              className="eui-alignBaseline"
            />
          </EuiLink>
        </Fragment>
      }
    >
      <EuiFlexGroup direction="column" gutterSize="s">
        {rules.length > 0 && (
          <EuiFlexItem key="default-route">
            <RuleCard
              isDefault={true}
              rule={default_traffic_rule}
              routes={routes}
              onChangeHandler={onChange("default_traffic_rule")}
              onDelete={null}
              errors={default_traffic_rule_errors}
            />
            <EuiHorizontalRule margin="xs" />
          </EuiFlexItem>
        )}
        <EuiDragDropContext onDragEnd={onDragEnd}>
          <EuiDroppable droppableId="CUSTOM_HANDLE_DROPPABLE_AREA" spacing="m">
            {rules.map((rule, idx) => (
              <EuiDraggable
                key={`${idx}`}
                index={idx}
                draggableId={`${idx}`}
                customDragHandle={true}
                disableInteractiveElementBlocking
              >
                {(provided) => (
                  <EuiFlexItem key={`${idx}`}>
                    <RuleCard
                      rule={rule}
                      routes={routes}
                      protocol={protocol}
                      onChangeHandler={onChange(`rules.${idx}`)}
                      onDelete={onDeleteRule(idx)}
                      errors={get(rules_errors, `${idx}`)}
                      dragHandleProps={provided.dragHandleProps}
                    />
                    <EuiSpacer size="s" />
                  </EuiFlexItem>
                )}
              </EuiDraggable>
            ))}
          </EuiDroppable>
        </EuiDragDropContext>
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
        <RulesPanelFlyout onClose={() => toggleIsFlyoutVisible()} />
      )}
    </Panel>
  );
};

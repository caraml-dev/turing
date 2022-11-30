import React, { useContext, useEffect, useMemo, useReducer } from "react";
import { EuiFlexItem, EuiFlexGroup, EuiSpacer, EuiText } from "@elastic/eui";
import { Panel } from "../../../Panel";
import ExperimentContext from "../../providers/context";
import { FormLabelWithToolTip } from "../../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { VariableConfigRow } from "./VariableConfigRow";
import { get } from "../../../../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../../../../components/form/hooks/useOnChangeHandler";
import { newVariableConfig } from "../../config";
import groupBy from "lodash/groupBy";
import sortBy from "lodash/sortBy";
import isArray from "lodash/isArray";
import isEqual from "lodash/isEqual";
import isEqualWith from "lodash/isEqualWith";

export const VariablesConfigPanel = ({
  variables,
  onChangeHandler,
  protocol,
  errors = {},
}) => {
  const {
    variables: allVariables,
    isLoaded,
    setVariablesValidated,
  } = useContext(ExperimentContext);

  // Update current variables config when new variables data arrives, if required
  useEffect(() => {
    if (isLoaded("variables")) {
      // Compare variables.config. Make a map of existing config variable names to the variables.
      const currentVars = variables.config.reduce((vars, v) => {
        vars[v.name] = v;
        return vars;
      }, {});
      // Check if the 2 variable lists are of the same length and contain the same var names
      const found = allVariables.config.reduce(
        (isFound, v) => isFound && v.name in currentVars,
        true
      );
      // Compare variables.client_variables and variables.experiment_variables, sorting arrays by name
      const varArrayComparator = (objValue, othValue) => {
        if (isArray(objValue) && isArray(othValue)) {
          return isEqual(sortBy(objValue, "name"), sortBy(othValue, "name"));
        }
      };
      if (
        !found ||
        allVariables.config.length !== variables.config.length ||
        !isEqualWith(
          allVariables.client_variables,
          variables.client_variables,
          varArrayComparator
        ) ||
        !isEqualWith(
          allVariables.experiment_variables,
          variables.experiment_variables,
          varArrayComparator
        )
      ) {
        onChangeHandler({
          ...allVariables,
          config: allVariables.config.map((v) => currentVars[v.name] || v),
        });
      }
      setVariablesValidated(true);
    }
  }, [
    variables,
    allVariables,
    isLoaded,
    setVariablesValidated,
    onChangeHandler,
  ]);

  const sortedVariables = useMemo(() => {
    // Assign a positional id to each element of variables.config for onChange handler.
    // Changing variables.config directly instead of mapping values to retain original reference to
    // the individual variables within the sortedVariables list. Set selected if field is not empty.
    variables.config.forEach((v, idx) => {
      v.idx = idx;
      !!v.field && (v.selected = true);
    });
    // Sort required variables by name, and append the remaining variables
    return [
      ...sortBy(
        variables.config.filter((v) => v.required),
        "name"
      ),
      ...variables.config.filter((v) => !v.required),
    ];
  }, [variables.config]);

  // Get the required variables and the optional variables whose values have been set
  const { selectedVariables = [], availableVariables = [] } = groupBy(
    sortedVariables,
    (v) => {
      return v.required || v.field || v.selected
        ? "selectedVariables"
        : "availableVariables";
    }
  );

  const { onChange } = useOnChangeHandler(onChangeHandler);
  // forceUpdate will be used to update the Variables UI on add/remove variable configurations.
  // eslint-disable-next-line no-unused-vars
  const [ignored, forceUpdate] = useReducer((x) => x + 1, 0);

  // Add a dummy variable to the selected variables
  !!availableVariables.length && selectedVariables.push(newVariableConfig());

  return (
    <Panel
      title={
        <FormLabelWithToolTip
          label="Variables"
          size="m"
          content="Specify how the required (and any optional) experiment variables may be parsed from the request."
        />
      }
    >
      <EuiSpacer size="xs" />
      {!!selectedVariables.length ? (
        <EuiFlexItem>
          <EuiFlexGroup direction="column" gutterSize="s">
            {selectedVariables.map((variable) => (
              <EuiFlexItem key={`experiment-variable-${variable.name}`}>
                <VariableConfigRow
                  variable={variable}
                  allVariables={sortedVariables}
                  availableVariables={availableVariables}
                  onChangeHandler={onChange(`config.${variable.idx}`)}
                  protocol={protocol}
                  forceUpdate={forceUpdate}
                  error={get(errors, `config.${variable.idx}`)}
                />
              </EuiFlexItem>
            ))}
          </EuiFlexGroup>
        </EuiFlexItem>
      ) : (
        <EuiText size="m" color="subdued">
          No variables available for configuration.
        </EuiText>
      )}
    </Panel>
  );
};

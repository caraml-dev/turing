import React, { useMemo, useState } from "react";
import {
  EuiButtonIcon,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSuperSelect,
} from "@elastic/eui";
import { useOnChangeHandler } from "../../../../../../../components/form/hooks/useOnChangeHandler";
import { FieldSourceFormLabel } from "../../../../../request_field_source/FieldSourceFormLabel";
import { resetVariableSelection } from "../../config";
import sortBy from "lodash/sortBy";
import "./VariableConfigRow.scss";

export const VariableConfigRow = ({
  variable,
  allVariables,
  availableVariables,
  onChangeHandler,
  forceUpdate,
  error = {},
}) => {
  // Add current variable to options list and sort by name
  const variableOptions = useMemo(() => {
    return sortBy([...availableVariables, { name: variable.name }], "name").map(
      (v) => ({
        value: v.name,
        inputDisplay: v.name,
      })
    );
  }, [availableVariables, variable.name]);

  // Save field source selection
  const [variableSource, setVariableSource] = useState(variable.field_source);

  // Define onChange handlers
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onDeleteVariableConfig = (name) => () => {
    const deleteVar = allVariables.find((v) => v.name === name);
    resetVariableSelection(deleteVar);
    forceUpdate();
  };

  const onChangeFieldSource = (fieldSource) => {
    setVariableSource(fieldSource);
    // If not a new variable, update the config
    !!variable.name && onChange("field_source")(fieldSource);
  };

  const onChangeName = (varName) => {
    // Update the new variable with the input configs from the current variable
    const newVariable = allVariables.find((v) => v.name === varName);
    if (!!newVariable) {
      newVariable.field = variable.field;
      newVariable.field_source = variableSource;
      // If !required and field value is currently unset, this indicates that the variable has been selected for configuration
      newVariable.selected = true;
      resetVariableSelection(variable);
      // Update the variables list
      const varNames = allVariables.map((v) => v.name);
      if (!!variable.name) {
        // If current variable name is not empty, swap position with new variable in the list
        allVariables[varNames.indexOf(variable.name)] = newVariable;
        allVariables[varNames.indexOf(newVariable.name)] = variable;
      } else {
        // Else, move new variable to the end of the list
        allVariables.splice(varNames.indexOf(newVariable.name), 1);
        allVariables.push(newVariable);
      }
      forceUpdate();
    }
  };

  return (
    <EuiFlexGroup
      direction="row"
      gutterSize="m"
      alignItems="center"
      className="euiFlexGroup--experimentVariableConfigRow">
      <EuiFlexItem grow={1} className="eui-textTruncate">
        <EuiFormRow isInvalid={!!error.name} error={error.name}>
          <EuiSuperSelect
            fullWidth
            compressed
            hasDividers={true}
            disabled={variable.required}
            valueOfSelected={variable.name}
            options={variableOptions}
            onChange={onChangeName}
            isInvalid={!!error.name}
          />
        </EuiFormRow>
      </EuiFlexItem>

      <EuiFlexItem grow={1} className="eui-textTruncate">
        <EuiFormRow isInvalid={!!error.field} error={error.field}>
          <EuiFieldText
            fullWidth
            compressed
            disabled={!variable.name}
            placeholder={"Enter Field Name..."}
            value={variable.field}
            onChange={(e) => onChange("field")(e.target.value)}
            isInvalid={!!error.field}
            prepend={
              <FieldSourceFormLabel
                readOnly={false}
                value={variableSource}
                onChange={onChangeFieldSource}
              />
            }
          />
        </EuiFormRow>
      </EuiFlexItem>

      <EuiFlexItem grow={false} className="euiFlexItem--hasActions">
        <EuiButtonIcon
          size="s"
          color="danger"
          iconType="trash"
          isDisabled={variable.required || !variable.name}
          onClick={onDeleteVariableConfig(variable.name)}
          aria-label="Remove Request Variable"
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

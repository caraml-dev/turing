import React, { useMemo } from "react";
import {
  EuiComboBox,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
} from "@elastic/eui";
import { get } from "../../../components/form/utils";
import { FieldSourceFormLabel } from "../request_field_source/FieldSourceFormLabel";

export const TrafficRuleCondition = ({
  condition,
  onChangeHandler,
  errors,
  readOnly,
}) => {
  const selectedOptions = useMemo(() => {
    return condition.values.map((v) => ({ label: v }));
  }, [condition.values]);

  const onChange = (field) => (value) => {
    condition[field] = value;
    onChangeHandler({ ...condition });
  };

  const onAddValue = (searchValue) => {
    onChange("values")([...condition.values, searchValue]);
  };

  return (
    <EuiFlexGroup direction="row" gutterSize="m" alignItems="baseline">
      <EuiFlexItem grow={1}>
        <EuiFormRow
          isInvalid={!!get(errors, "field")}
          error={get(errors, "field")}
        >
          <EuiFieldText
            readOnly={readOnly}
            compressed
            value={condition.field || ""}
            onChange={(e) => onChange("field")(e.target.value)}
            isInvalid={!!get(errors, "field")}
            prepend={
              <FieldSourceFormLabel
                readOnly={readOnly}
                value={condition.field_source}
                onChange={onChange("field_source")}
              />
            }
          />
        </EuiFormRow>
      </EuiFlexItem>

      <EuiFlexItem grow={false}>{condition.operator}</EuiFlexItem>

      <EuiFlexItem grow={1}>
        <EuiFormRow
          isInvalid={!!get(errors, "values")}
          error={get(errors, "values")}
        >
          <EuiComboBox
            compressed
            isDisabled={readOnly}
            customOptionText="Add {searchValue} as a value"
            selectedOptions={selectedOptions}
            onCreateOption={onAddValue}
            onChange={(selected) =>
              onChange("values")(selected.map((l) => l.label))
            }
            isInvalid={!!get(errors, "values")}
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

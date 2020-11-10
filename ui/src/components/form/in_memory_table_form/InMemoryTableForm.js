import React, { useMemo } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiInMemoryTable
} from "@elastic/eui";
import { get } from "../utils";

import "./InMemoryTableForm.scss";

export const InMemoryTableForm = ({
  errors = {},
  renderErrorHeader = key => key,
  ...props
}) => {
  const errorMessage = useMemo(() => {
    return Object.keys(errors).flatMap(fieldName => {
      const fieldErrors = Object.keys(errors[fieldName]).flatMap(
        key => get(errors[fieldName], key) || []
      );
      return (
        <EuiFlexGroup direction="row" gutterSize="s">
          <EuiFlexItem grow={false}>
            <strong>{renderErrorHeader(fieldName)}:</strong>
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiFlexGroup direction="column" gutterSize="none">
              {fieldErrors.map((error, idx) => (
                <EuiFlexItem grow={false} key={idx}>
                  {error}
                </EuiFlexItem>
              ))}
            </EuiFlexGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      );
    });
  }, [errors, renderErrorHeader]);

  return (
    <EuiFormRow fullWidth isInvalid={!!errorMessage} error={errorMessage}>
      <EuiInMemoryTable
        className="euiBasicTable--inMemoryFormTable"
        {...props}
      />
    </EuiFormRow>
  );
};

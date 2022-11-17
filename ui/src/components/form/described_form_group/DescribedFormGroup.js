import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiText } from "@elastic/eui";
import "./DescribedFormGroup.scss";

export const DescribedFormGroup = ({ description, ...props }) => (
  <EuiFlexGroup direction="row" gutterSize="none">
    <EuiFlexItem grow={1}>{props.children}</EuiFlexItem>

    {description && (
      <EuiFlexItem grow={2} className="euiFlexItem--engineTypeDescription">
        <EuiText size="s" color="subdued">
          {description}
        </EuiText>
      </EuiFlexItem>
    )}
  </EuiFlexGroup>
);

import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiHighlight, EuiIcon } from "@elastic/eui";

export const EuiComboboxSuggestItemRowHeight = 40;

export const ComboboxSuggestItem = ({
  option,
  searchValue,
  optionContentClassname,
  prepend,
  append,
}) => {
  return (
    <EuiFlexGroup direction="row" justifyContent="spaceBetween">
      {append ? (
        <EuiFlexItem grow={false}>{append}</EuiFlexItem>
      ) : (
        option.icon && (
          <EuiFlexItem grow={false} style={{ marginRight: 0 }}>
            <EuiIcon type={option.icon} style={{ margin: "auto" }} />
          </EuiFlexItem>
        )
      )}

      <EuiFlexItem grow={4}>
        <EuiHighlight search={searchValue} className={optionContentClassname}>
          {option.label}
        </EuiHighlight>
      </EuiFlexItem>

      {prepend && (
        <EuiFlexItem className="comboboxOption__suggest--prepend">
          {prepend}
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};

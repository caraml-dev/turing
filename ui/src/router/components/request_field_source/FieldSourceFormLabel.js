import React, { useMemo } from "react";
import { flattenPanelTree, useToggle } from "@gojek/mlp-ui";
import {
  EuiButtonEmpty,
  EuiContextMenu,
  EuiFormLabel,
  EuiPopover
} from "@elastic/eui";

import "./FieldSourceFormLabel.scss";

const fieldSourceOptions = [
  {
    value: "header",
    inputDisplay: "Header"
  },
  {
    value: "payload",
    inputDisplay: "Payload"
  }
];

export const FieldSourceFormLabel = ({ value, onChange, readOnly }) => {
  const [isOpen, togglePopover] = useToggle();

  const panels = flattenPanelTree({
    id: 0,
    items: fieldSourceOptions.map(option => ({
      name: option.inputDisplay,
      value: option.value,
      onClick: () => {
        togglePopover();
        onChange(option.value);
      }
    }))
  });

  const selectedOption = useMemo(
    () => fieldSourceOptions.find(o => o.value === value),
    [value]
  );

  return readOnly ? (
    <EuiFormLabel className="fieldSourceLabel euiFormControlLayout__prepend">
      {selectedOption.inputDisplay}
    </EuiFormLabel>
  ) : (
    <EuiPopover
      button={
        <EuiButtonEmpty
          size="xs"
          iconType="arrowDown"
          iconSide="right"
          className="fieldSourceLabel"
          onClick={togglePopover}>
          {selectedOption.inputDisplay}
        </EuiButtonEmpty>
      }
      isOpen={isOpen}
      closePopover={togglePopover}
      panelPaddingSize="s">
      <EuiContextMenu
        className="fieldSourceDropdown"
        initialPanelId={0}
        panels={panels}
      />
    </EuiPopover>
  );
};

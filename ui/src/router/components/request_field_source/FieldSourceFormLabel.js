import React, { useMemo } from "react";
import { flattenPanelTree, useToggle } from "@caraml-dev/ui-lib";
import {
  EuiButtonEmpty,
  EuiContextMenu,
  EuiFormLabel,
  EuiPopover,
} from "@elastic/eui";

import "./FieldSourceFormLabel.scss";

const fieldSourceOptions = {
  HTTP_JSON: [
    {
      value: "header",
      inputDisplay: "Header",
    },
    {
      value: "payload",
      inputDisplay: "Payload",
    },
  ],
  UPI_V1: [
    {
      value: "header",
      inputDisplay: "Header",
    },
    {
      value: "prediction_context",
      inputDisplay: "Prediction Context",
      toolTipContent: "Prediction Context of the UPI proto",
    },
  ],
};

export const FieldSourceFormLabel = ({
  value,
  onChange,
  readOnly,
  protocol,
}) => {
  const [isOpen, togglePopover] = useToggle();
  const options = fieldSourceOptions[protocol];

  const panels = flattenPanelTree({
    id: 0,
    items: options.map((option) => ({
      name: option.inputDisplay,
      value: option.value,
      toolTipContent: option.toolTipContent,
      toolTipPosition: "right",
      onClick: () => {
        togglePopover();
        onChange(option.value);
      },
    })),
  });

  const selectedOption = useMemo(
    () => fieldSourceOptions[protocol].find((o) => o.value === value),
    [value, protocol]
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
          onClick={togglePopover}
        >
          {selectedOption.inputDisplay}
        </EuiButtonEmpty>
      }
      isOpen={isOpen}
      closePopover={togglePopover}
      panelPaddingSize="s"
    >
      <EuiContextMenu
        className="fieldSourceDropdown"
        initialPanelId={0}
        panels={panels}
      />
    </EuiPopover>
  );
};

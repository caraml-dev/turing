import React, { useState } from "react";
import { EuiButtonEmpty, EuiContextMenu, EuiPopover } from "@elastic/eui";
import { flattenPanelTree } from "@gojek/mlp-ui";

export const DockerRegistryPopover = ({ value, registryOptions, onChange }) => {
  const [isOpen, setOpen] = useState(false);

  const panels = flattenPanelTree({
    id: 0,
    items: registryOptions.map((registry) => ({
      name: registry.inputDisplay,
      value: registry.value,
      icon: "logoDocker",
      onClick: () => {
        togglePopover();
        onChange(registry.value);
      },
    })),
  });

  const togglePopover = () => setOpen(!isOpen);

  return (
    <EuiPopover
      button={
        <EuiButtonEmpty
          size="xs"
          iconType="arrowDown"
          iconSide="right"
          onClick={togglePopover}
        >
          {registryOptions.find((o) => o.value === value).inputDisplay}
        </EuiButtonEmpty>
      }
      isOpen={isOpen}
      closePopover={togglePopover}
      panelPaddingSize="s"
    >
      <EuiContextMenu initialPanelId={0} panels={panels} />
    </EuiPopover>
  );
};

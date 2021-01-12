import React from "react";
import {
  EuiButtonEmpty,
  EuiButtonIcon,
  EuiScreenReaderOnly,
  EuiToolTip,
  htmlIdGenerator
} from "@elastic/eui";
import isString from "lodash/isString";

export const DefaultItemAction = ({ action, enabled, item, className }) => {
  if (!action.onClick && !action.href) {
    throw new Error(`Cannot render item action [${action.name}]. Missing required 'onClick' callback
      or 'href' string. If you want to provide a custom action control, make sure to define the 'render' callback`);
  }

  const onClick = action.onClick ? () => action.onClick(item) : undefined;

  const buttonColor = action.color;
  let color = "primary";
  if (buttonColor) {
    color = isString(buttonColor) ? buttonColor : buttonColor(item);
  }

  const buttonIcon = action.icon;
  let icon;
  if (buttonIcon) {
    icon = isString(buttonIcon) ? buttonIcon : buttonIcon(item);
  }

  let button;
  const actionContent =
    typeof action.name === "function" ? action.name(item) : action.name;
  if (action.type === "icon") {
    if (!icon) {
      throw new Error(`Cannot render item action [${action.name}]. It is configured to render as an icon but no
      icon is provided. Make sure to set the 'icon' property of the action`);
    }
    const ariaLabelId = htmlIdGenerator()();
    button = (
      <>
        <EuiButtonIcon
          className={className}
          aria-labelledby={ariaLabelId}
          isDisabled={!enabled}
          color={color}
          iconType={icon}
          onClick={onClick}
          href={action.href}
          target={action.target}
          data-test-subj={action["data-test-subj"]}
        />
        {/* actionContent (action.name) is a ReactNode and must be rendered to an element and referenced by ID for screen readers */}
        <EuiScreenReaderOnly>
          <span id={ariaLabelId}>{actionContent}</span>
        </EuiScreenReaderOnly>
      </>
    );
  } else {
    button = (
      <EuiButtonEmpty
        className={className}
        size={action.size}
        isDisabled={!enabled}
        color={color}
        iconType={icon}
        onClick={onClick}
        href={action.href}
        target={action.target}
        data-test-subj={action["data-test-subj"]}
        flush="right">
        {actionContent}
      </EuiButtonEmpty>
    );
  }

  return enabled && action.description ? (
    <EuiToolTip content={action.description} delay="long">
      {button}
    </EuiToolTip>
  ) : (
    button
  );
};

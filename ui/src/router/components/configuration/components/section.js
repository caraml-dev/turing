import React, { Fragment } from "react";
import {
  EuiHorizontalRule,
  EuiIcon,
  EuiPanel,
  EuiTextColor,
  EuiTitle
} from "@elastic/eui";
import "./section.scss";

export const ConfigSectionTitle = ({ title, iconType }) => (
  <EuiTitle size="s">
    <EuiTextColor color="secondary">
      <span>
        {!!iconType && (
          <EuiIcon className="eui-alignBaseline" type={iconType} size="m" />
        )}
        &nbsp;{title}
      </span>
    </EuiTextColor>
  </EuiTitle>
);

export const ConfigSection = ({ title, iconType, children }) => (
  <Fragment>
    <ConfigSectionTitle title={title} iconType={iconType} />
    {children}
  </Fragment>
);

export const ConfigSectionPanelTitle = ({ title }) => (
  <Fragment>
    <EuiTitle size="xs">
      <span>{title}</span>
    </EuiTitle>
    <EuiHorizontalRule size="full" margin="xs" />
  </Fragment>
);

export const ConfigSectionPanel = React.forwardRef(
  ({ title, ...props }, ref) => (
    <EuiPanel className={`euiPanel--configSection ${props.className}`}>
      <div ref={ref}>
        {title && <ConfigSectionPanelTitle title={title} />}
        {props.children}
      </div>
    </EuiPanel>
  )
);

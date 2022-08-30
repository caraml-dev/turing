import React from "react";
import { EuiIcon, EuiTextColor, EuiTitle, EuiSpacer } from "@elastic/eui";

export const ConfigSectionTitle = ({ title, iconType }) => (
  <EuiTitle size="s">
    <EuiTextColor color="#00BFB3">
      <span>
        {!!iconType && (
          <EuiIcon className="eui-alignBaseline" type={iconType} size="m" />
        )}
        &nbsp;{title}
      </span>
      <EuiSpacer size="s" />
    </EuiTextColor>
  </EuiTitle>
);

import React from "react";
import { EuiIcon, EuiTextColor, EuiTitle } from "@elastic/eui";

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

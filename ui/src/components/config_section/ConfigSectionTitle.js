import React from "react";
import { EuiIcon, EuiTextColor, EuiTitle, EuiSpacer } from "@elastic/eui";

export const ConfigSectionTitle = ({ title, iconType }) => (
  <EuiTitle size="s">
    <span>
      <EuiTextColor color="success">
        <EuiSpacer size="s" />
          <span>
            {!!iconType && (
              <EuiIcon className="eui-alignBaseline" type={iconType} size="m" />
            )}
            &nbsp;{title}
          </span>
        <EuiSpacer size="s" />
      </EuiTextColor>
    </span>
  </EuiTitle>
);

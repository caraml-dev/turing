import { EuiIcon, EuiTitle } from "@elastic/eui";
import { appConfig } from "../../config";
import React from "react";

export const PageTitle = ({ title, size }) => (
  <EuiTitle size={size || "l"}>
    <span>
      <EuiIcon type={appConfig.appIcon} size="xl" />
      &nbsp;{title}
    </span>
  </EuiTitle>
);

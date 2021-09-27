import React, { Fragment } from "react";
import { EuiHorizontalRule, EuiTitle } from "@elastic/eui";

export const ConfigSectionPanelTitle = ({ title }) => (
  <Fragment>
    <EuiTitle size="xs">
      <span>{title}</span>
    </EuiTitle>
    <EuiHorizontalRule size="full" margin="xs" />
  </Fragment>
);

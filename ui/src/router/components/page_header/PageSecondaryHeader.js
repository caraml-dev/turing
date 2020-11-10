import React from "react";
import { EuiPanel } from "@elastic/eui";

import "./PageSecondaryHeader.scss";

export const PageSecondaryHeader = props => {
  return <EuiPanel className="euiPanel--subHeader">{props.children}</EuiPanel>;
};

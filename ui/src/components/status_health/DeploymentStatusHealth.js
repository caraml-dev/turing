import React from "react";
import { EuiHealth } from "@elastic/eui";

export const DeploymentStatusHealth = ({ status }) => {
  return <EuiHealth color={status.color}>{status.label}</EuiHealth>;
};

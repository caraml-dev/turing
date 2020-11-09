import React from "react";
import { EuiHealth } from "@elastic/eui";
import { Status } from "../../../services/status/Status";

export const DeploymentStatusHealth = ({ status: statusStr }) => {
  const status = Status.fromValue(statusStr);
  return <EuiHealth color={status.color}>{status.label}</EuiHealth>;
};

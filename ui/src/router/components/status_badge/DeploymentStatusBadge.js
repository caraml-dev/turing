import React from "react";
import { EuiBadge } from "@elastic/eui";
import "./DeploymentStatusBadge.scss";
import { Status } from "../../../services/status/Status";

export const DeploymentStatusBadge = ({ status: statusStr }) => {
  const status = Status.fromValue(statusStr);
  return (
    <EuiBadge
      className="euiBadge--status"
      color={status.color}
      iconType={status.iconType}>
      {status.label}
    </EuiBadge>
  );
};

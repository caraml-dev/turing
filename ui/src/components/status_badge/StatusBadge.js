import React from "react";
import { EuiBadge } from "@elastic/eui";
import "./StatusBadge.scss";

export const StatusBadge = ({ status }) =>
  !!status ? (
    <EuiBadge
      className="euiBadge--status"
      color={status.color}
      iconType={status.iconType}
    >
      {status.label}
    </EuiBadge>
  ) : null;

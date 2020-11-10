import React from "react";
import { EuiIcon, EuiToolTip } from "@elastic/eui";

export const FormLabelWithToolTip = ({ label, content, size = "s" }) => (
  <EuiToolTip content={content}>
    <span>
      {label}&nbsp;
      <EuiIcon
        size={size}
        type="questionInCircle"
        color="subdued"
        className="eui-alignBaseline"
      />
    </span>
  </EuiToolTip>
);

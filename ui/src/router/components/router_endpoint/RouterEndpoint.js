import { EuiCopy, EuiIcon, EuiLink, EuiText, EuiTextColor } from "@elastic/eui";
import React from "react";

export const RouterEndpoint = ({ endpoint, className }) => {
  return endpoint ? (
    <EuiCopy
      textToCopy={endpoint}
      beforeMessage="Click to copy URL to clipboard">
      {(copy) => (
        <EuiLink
          className={className}
          onClick={(e) => {
            e.stopPropagation();
            copy();
          }}
          color="text">
          <EuiText size="s">
            <EuiIcon
              type={"copyClipboard"}
              size="s"
              style={{ marginRight: "inherit" }}
            />
            &nbsp;{endpoint}
          </EuiText>
        </EuiLink>
      )}
    </EuiCopy>
  ) : (
    <EuiTextColor color="subdued">Not available</EuiTextColor>
  );
};

import React from "react";

import "./OverlayMask.scss";
import useDimension from "../../hooks/useDimension";

export const OverlayMask = ({ parentRef, opacity = 0.5, children }) => {
  const { width, height } = useDimension(parentRef);

  return (
    <div
      className="overlayMask"
      style={{ width, height, background: `rgba(255, 255, 255, ${opacity})` }}>
      {children}
    </div>
  );
};

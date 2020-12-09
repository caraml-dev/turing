import React, { Fragment, useRef } from "react";
import { EuiButton, EuiFlexItem, EuiLink, EuiSpacer } from "@elastic/eui";
import classNames from "classnames";
import { useToggle } from "@gojek/mlp-ui";

import "./ExpandableDescriptionList.scss";
import useDimension from "../../hooks/useDimension";

export const ExpandableContainer = ({
  maxHeight = 300,
  isScrollable = true,
  toggleKind = "button", // "button" | "text"
  children
}) => {
  const contentRef = useRef();
  const [isExpanded, toggle] = useToggle();
  const { height: contentHeight } = useDimension(contentRef);

  return (
    <div
      className={classNames("expandableContainer", {
        "expandableContainer-isOpen": isExpanded
      })}>
      <EuiFlexItem grow={true} className="euiFlexItem--childFlexPanel">
        <div
          className={classNames("expandableContainer__childWrapper", {
            scrollableContainer: isScrollable
          })}
          style={{
            height: isExpanded
              ? contentHeight
              : Math.min(contentHeight, maxHeight)
          }}>
          <div ref={contentRef}>{children}</div>
        </div>
      </EuiFlexItem>

      {contentHeight > maxHeight && (
        <Fragment>
          <EuiSpacer size="s" />
          <div className="expandableContainer__triggerWrapper">
            {toggleKind === "button" ? (
              <EuiButton
                fullWidth
                onClick={toggle}
                iconType={`arrow${isExpanded ? "Up" : "Right"}`}>
                {isExpanded ? "Collapse" : "Expand"}
              </EuiButton>
            ) : (
              <EuiLink type="button" onClick={toggle}>
                {isExpanded ? "Show less" : "Show more"}
              </EuiLink>
            )}
          </div>
        </Fragment>
      )}
    </div>
  );
};

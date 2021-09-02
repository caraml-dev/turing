import React, { useRef } from "react";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiLink } from "@elastic/eui";
import classNames from "classnames";
import { useToggle } from "@gojek/mlp-ui";
import useDimension from "../../hooks/useDimension";
import "./ExpandableContainer.scss";

export const ExpandableContainer = ({
  maxCollapsedHeight = 300,
  isScrollable = true,
  toggleKind = "button", // "button" | "text"
  children,
}) => {
  const contentRef = useRef();
  const [isExpanded, toggle] = useToggle();
  const { height: contentHeight } = useDimension(contentRef);

  const toggleComponent =
    toggleKind === "button" ? (
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
    );

  return (
    <EuiFlexGroup
      direction="column"
      gutterSize="xs"
      className={classNames("expandableContainer", {
        "expandableContainer-isOpen": isExpanded,
      })}>
      <EuiFlexItem grow={true}>
        <div
          className={classNames("expandableContainer__childWrapper", {
            scrollableContainer: isScrollable,
          })}
          style={{
            height: isExpanded
              ? contentHeight
              : Math.min(contentHeight, maxCollapsedHeight),
          }}>
          <div ref={contentRef}>{children}</div>
        </div>
      </EuiFlexItem>

      {contentHeight > maxCollapsedHeight && (
        <EuiFlexItem>{toggleComponent}</EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};

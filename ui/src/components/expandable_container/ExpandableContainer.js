import React from "react";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiLink } from "@elastic/eui";
import classNames from "classnames";
import useCollapse from "react-collapsed";
import useDimensions from "react-use-dimensions";
import "./ExpandableContainer.scss";

const Content = React.forwardRef(({ children }, ref) => (
  <div ref={ref}>{children}</div>
));

const ExpandableContainer = ({
  collapsedHeight,
  isScrollable,
  toggleKind,
  children,
}) => {
  const { getCollapseProps, getToggleProps, isExpanded } = useCollapse({
    collapsedHeight,
  });

  const toggle =
    toggleKind === "button" ? (
      <EuiButton
        fullWidth
        {...getToggleProps()}
        iconType={`arrow${isExpanded ? "Up" : "Right"}`}
      >
        {isExpanded ? "Collapse" : "Expand"}
      </EuiButton>
    ) : (
      <EuiLink {...getToggleProps()}>
        {isExpanded ? "Show less" : "Show more"}
      </EuiLink>
    );

  return (
    <EuiFlexGroup direction="column" gutterSize="xs">
      <EuiFlexItem grow={true}>
        <div
          className={classNames({ scrollableContainer: isScrollable })}
          {...getCollapseProps()}
        >
          {children}
        </div>
      </EuiFlexItem>

      <EuiFlexItem>{toggle}</EuiFlexItem>
    </EuiFlexGroup>
  );
};

const ExpandableContainerWrapper = ({
  maxCollapsedHeight = 300,
  isScrollable = true,
  toggleKind = "button", // "button" | "text"
  children,
}) => {
  const [contentRef, { height: contentHeight }] = useDimensions({
    liveMeasure: false,
  });

  const content = <Content ref={contentRef}>{children}</Content>;

  return contentHeight > maxCollapsedHeight ? (
    <ExpandableContainer
      collapsedHeight={maxCollapsedHeight}
      isScrollable={isScrollable}
      toggleKind={toggleKind}
    >
      {content}
    </ExpandableContainer>
  ) : (
    content
  );
};

export { ExpandableContainerWrapper as ExpandableContainer };

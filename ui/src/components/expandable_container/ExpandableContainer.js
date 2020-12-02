import React, { Fragment, useRef } from "react";
import {
  EuiButton,
  EuiButtonIcon,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer
} from "@elastic/eui";
import classNames from "classnames";
import { useToggle } from "@gojek/mlp-ui";

import "./ExpandableDescriptionList.scss";
import useDimension from "../../hooks/useDimension";

export const ExpandableContainer = ({
  maxHeight = 300,
  buttonSize = "l",
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
      <EuiFlexGroup direction="row" gutterSize="xs" alignItems="left">
        <EuiFlexItem grow={true} className="euiFlexItem--childFlexPanel">
          <div
            className="expandableContainer__childWrapper"
            style={{
              height: isExpanded
                ? contentHeight
                : Math.min(contentHeight, maxHeight)
            }}>
            <div ref={contentRef}>{children}</div>
          </div>
        </EuiFlexItem>

        {contentHeight > maxHeight && buttonSize === "s" && (
          <EuiFlexItem grow={false}>
            <EuiButtonIcon
              onClick={toggle}
              iconType={`arrow${isExpanded ? "Up" : "Right"}`}
              aria-label="Next"
            />
          </EuiFlexItem>
        )}
      </EuiFlexGroup>

      {contentHeight > maxHeight && buttonSize === "l" && (
        <Fragment>
          <EuiSpacer size="s" />
          <div className="expandableContainer__triggerWrapper">
            <EuiButton
              fullWidth
              onClick={toggle}
              iconType={`arrow${isExpanded ? "Up" : "Right"}`}>
              {isExpanded ? "Collapse" : "Expand"}
            </EuiButton>
          </div>
        </Fragment>
      )}
    </div>
  );
};

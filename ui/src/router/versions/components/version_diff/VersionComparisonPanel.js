import React, { useState } from "react";
import {
  EuiFilterButton,
  EuiFilterGroup,
  EuiFlexGroup,
  EuiFlexItem,
} from "@elastic/eui";
import ReactDiffViewer, { DiffMethod } from "react-diff-viewer";

const diffViewerStyles = {
  line: {
    wordBreak: "break-word",
    fontSize: "0.775rem",
  },
};

export const VersionComparisonPanel = ({
  leftValue,
  rightValue,
  leftTitle,
  rightTitle,
}) => {
  const [splitView, setSplitView] = useState(true);

  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      <EuiFlexItem>
        <EuiFlexGroup direction="row" justifyContent="flexEnd">
          <EuiFlexItem grow={false}>
            <EuiFilterGroup>
              <EuiFilterButton
                hasActiveFilters={!splitView}
                onClick={() => setSplitView(false)}
              >
                Unified
              </EuiFilterButton>
              <EuiFilterButton
                hasActiveFilters={splitView}
                onClick={() => setSplitView(true)}
              >
                Split
              </EuiFilterButton>
            </EuiFilterGroup>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>

      <EuiFlexItem>
        <ReactDiffViewer
          leftTitle={leftTitle}
          rightTitle={rightTitle}
          oldValue={leftValue}
          newValue={rightValue}
          styles={diffViewerStyles}
          compareMethod={DiffMethod.WORDS_WITH_SPACE}
          splitView={splitView}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

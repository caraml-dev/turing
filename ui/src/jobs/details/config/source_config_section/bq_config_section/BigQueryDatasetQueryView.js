import React from "react";
import { EuiCodeBlock } from "@elastic/eui";
import dedent from "ts-dedent";

export const BigQueryDatasetQueryView = ({ query, maxHeight = 300 }) => (
  <EuiCodeBlock
    language="sql"
    fontSize="m"
    paddingSize="m"
    overflowHeight={maxHeight}
    isCopyable
  >
    {dedent(query)}
  </EuiCodeBlock>
);

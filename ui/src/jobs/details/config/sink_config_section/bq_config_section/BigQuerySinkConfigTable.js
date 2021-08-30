import React from "react";
import { getBigQueryConsoleUrl } from "../../../../../utils/gcp";
import {
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";

export const BigQuerySinkConfigTable = ({
  bqConfig: { table, staging_bucket, options },
}) => {
  const items = [
    {
      title: "Destination Table",
      description: (
        <EuiLink href={getBigQueryConsoleUrl(table)} target="_blank">
          {table}
        </EuiLink>
      ),
    },
    {
      title: "GCS Staging Bucket",
      description: staging_bucket,
    },
    // {
    //   title: "Extra Options",
    //   description: (
    //     <EuiCodeBlock
    //       language="json"
    //       fontSize="m"
    //       transparentBackground
    //       paddingSize="m">
    //       {JSON.stringify(options, null, 2)}
    //     </EuiCodeBlock>
    //   )
    // }
  ];

  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      <EuiFlexItem>
        <EuiDescriptionList
          compressed
          textStyle="reverse"
          type="responsiveColumn"
          listItems={items}
          titleProps={{ style: { width: "30%" } }}
          descriptionProps={{ style: { width: "70%" } }}
        />
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiTitle size="xxs">
          <EuiTextColor color="secondary">Extra Options</EuiTextColor>
        </EuiTitle>
        {Object.entries(options).length ? (
          <EuiDescriptionList
            compressed
            textStyle="reverse"
            type="responsiveColumn"
            listItems={Object.entries(options).map(([title, description]) => ({
              title,
              description,
            }))}
            titleProps={{ style: { width: "30%" } }}
            descriptionProps={{ style: { width: "70%" } }}
          />
        ) : (
          "None"
        )}
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

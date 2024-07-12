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
import { BigQueryOptionsTable } from "../../components/bq_options_table/BigQueryOptionsTable";

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
  ];

  return (
    <EuiFlexGroup direction="column" gutterSize="s">
      <EuiFlexItem>
        <EuiDescriptionList
          compressed
          textStyle="reverse"
          type="responsiveColumn"
          listItems={items}
          columnWidths={[1, 7/3]}
        />
      </EuiFlexItem>

      <EuiFlexItem>
        <EuiTitle size="xxs">
          <span>
            <EuiTextColor color="success">Extra Options</EuiTextColor>
          </span>
        </EuiTitle>

        <BigQueryOptionsTable options={options} />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

import React, { Fragment } from "react";
import {
  EuiAccordion,
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
  EuiSpacer,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import { BigQueryDatasetQueryView } from "./BigQueryDatasetQueryView";
import { CollapsibleBadgeGroup } from "../../../../../components/collapsible_badge_group/CollapsibleBadgeGroup";
import { BigQueryOptionsTable } from "../../components/bq_options_table/BigQueryOptionsTable";
import { ConfigSectionPanel } from "../../../../../components/config_section";
import { slugify } from "@gojek/mlp-ui";

import "./BigQueryDatasetSection.scss";
import { getBigQueryConsoleUrl } from "../../../../../utils/gcp";

const AccordionItem = ({ title, ...props }) => (
  <EuiAccordion
    paddingSize="xs"
    id={slugify(title)}
    buttonContent={
      <EuiTitle size="xxs">
        <span>
          <EuiTextColor color="success">{title}</EuiTextColor>
        </span>
      </EuiTitle>
    }>
    <div className="childWrapper">{props.children}</div>
  </EuiAccordion>
);

const BigQueryDatasetTableConfigGroup = ({ table, features, options }) => {
  const items = [
    {
      title: "Source Table",
      description: (
        <EuiLink href={getBigQueryConsoleUrl(table)} target="_blank">
          {table}
        </EuiLink>
      ),
    },
    {
      title: "Features",
      description: (
        <CollapsibleBadgeGroup
          className="euiBadgeGroup---insideDescriptionList"
          items={features}
          minItemsCount={5}
        />
      ),
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
          titleProps={{ style: { width: "160px" } }}
          descriptionProps={{ style: { width: "calc(100% - 160px)" } }}
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

export const BigQueryDatasetSectionPanel = ({
  bqConfig: { table, query, features, options },
}) => {
  return (
    <ConfigSectionPanel title="BigQuery Dataset">
      {!!table && (
        <BigQueryDatasetTableConfigGroup
          table={table}
          features={features}
          options={options}
        />
      )}

      {!!query && (
        <Fragment>
          <AccordionItem title="Query">
            <BigQueryDatasetQueryView query={query} />
          </AccordionItem>
          <EuiSpacer size="xs" />

          <AccordionItem title="Extra Options">
            <BigQueryOptionsTable options={options} />
          </AccordionItem>
        </Fragment>
      )}
    </ConfigSectionPanel>
  );
};

import React from "react";
import { EuiLink } from "@elastic/eui";
import { HorizontalDescriptionList } from "../../../../components/horizontal_description_list/HorizontalDescriptionList";
import { DateFromNow } from "@gojek/mlp-ui";
import { PageSecondaryHeader } from "../../../components/page_header/PageSecondaryHeader";
import { get } from "../../../../components/form/utils";
import { DeploymentStatusHealth } from "../../../../components/status_health/DeploymentStatusHealth";

export const VersionDetailsPageHeader = ({ version }) => {
  const headerItems = [
    {
      title: "Router",
      description: (
        <EuiLink href="../../">{get(version, "router.name")}</EuiLink>
      ),
    },
    {
      title: "Status",
      description: <DeploymentStatusHealth status={version.status} />,
    },
    {
      title: "Created At",
      description: <DateFromNow date={version.created_at} size="s" />,
    },
    {
      title: "Updated At",
      description: <DateFromNow date={version.updated_at} size="s" />,
    },
  ];

  return (
    <PageSecondaryHeader>
      <HorizontalDescriptionList listItems={headerItems} />
    </PageSecondaryHeader>
  );
};

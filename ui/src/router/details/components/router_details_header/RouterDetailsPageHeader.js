import React from "react";
import { EuiBadge } from "@elastic/eui";
import { DateFromNow } from "@gojek/mlp-ui";
import { HorizontalDescriptionList } from "../../../../components/horizontal_description_list/HorizontalDescriptionList";
import { PageSecondaryHeader } from "../../../components/page_header/PageSecondaryHeader";
import { RouterEndpoint } from "../../../components/router_endpoint/RouterEndpoint";

import "./RouterDetailsPageHeader.scss";

export const RouterDetailsPageHeader = ({ router }) => {
  const headerItems = [
    {
      title: "Endpoint",
      description: (
        <RouterEndpoint
          endpoint={router.endpoint}
          className="eui-textTruncate"
        />
      ),
      flexProps: {
        grow: true,
        className: "euiFlexItem--endpoint",
      },
    },
    {
      title: "Environment",
      description: (
        <EuiBadge color="default">{router.environment_name}</EuiBadge>
      ),
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px",
        },
      },
    },
    {
      title: "Created At",
      description: <DateFromNow date={router.created_at} size="s" />,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px",
        },
      },
    },
    {
      title: "Updated At",
      description: <DateFromNow date={router.updated_at} size="s" />,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px",
        },
      },
    },
  ];

  return (
    <PageSecondaryHeader>
      <HorizontalDescriptionList listItems={headerItems} />
    </PageSecondaryHeader>
  );
};

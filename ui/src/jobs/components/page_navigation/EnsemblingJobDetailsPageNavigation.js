import React from "react";
import { EuiIcon } from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";

export const EnsemblingJobDetailsPageNavigation = ({ job, selectedTab }) => {
  const tabs = [
    {
      id: "details",
      name: "Configuration",
    },
    {
      id: "logs",
      name: "Logs",
    },
    {
      id: "monitoring",
      name: (
        <span>
          Monitoring&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </span>
      ),
      href: job.monitoring_url,
      disabled: !job.monitoring_url,
    },
  ];

  return <PageNavigation tabs={tabs} selectedTab={selectedTab} />;
};

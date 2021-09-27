import { EuiIcon } from "@elastic/eui";
import { PageNavigation } from "../../../components/page_navigation/PageNavigation";
import React from "react";

export const EnsemblingJobDetailsPageNavigation = ({ job, ...props }) => {
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

  return <PageNavigation tabs={tabs} selectedTab={props["*"]} {...props} />;
};

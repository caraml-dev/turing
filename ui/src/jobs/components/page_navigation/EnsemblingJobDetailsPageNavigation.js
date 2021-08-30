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
      disabled: true,
    },
    {
      id: "monitoring",
      name: (
        <span>
          Monitoring&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </span>
      ),
      href: "blank",
      disabled: true,
    },
  ];

  return <PageNavigation tabs={tabs} selectedTab={props["*"]} {...props} />;
};

import React from "react";
import { PageNavigation } from "../../../../components/page_navigation/PageNavigation";
import { EuiIcon } from "@elastic/eui";

export const RouterVersionDetailsPageNavigation = ({
  version,
  actions,
  ...props
}) => {
  const tabs = [
    {
      id: "details",
      name: "Configuration",
    },
    {
      id: "monitoring_dashboard_link",
      name: (
        <span>
          Monitoring&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </span>
      ),
      href: version.monitoring_url,
    },
  ];

  return (
    <PageNavigation
      tabs={tabs}
      actions={actions}
      selectedTab={props["*"]}
      {...props}
    />
  );
};

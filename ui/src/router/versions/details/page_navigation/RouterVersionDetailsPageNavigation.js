import React from "react";
import { EuiIcon } from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";

export const RouterVersionDetailsPageNavigation = ({
  version,
  actions,
  selectedTab,
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
      selectedTab={selectedTab}
    />
  );
};

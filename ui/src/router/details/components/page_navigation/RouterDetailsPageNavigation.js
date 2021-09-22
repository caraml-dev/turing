import React from "react";
import { EuiIcon } from "@elastic/eui";
import { PageNavigation } from "../../../../components/page_navigation/PageNavigation";

export const RouterDetailsPageNavigation = ({
  router: { config = {}, ...router },
  actions,
  ...props
}) => {
  const tabs = [
    {
      id: "details",
      name: "Configuration",
    },
    {
      id: "history",
      name: "History",
    },
    {
      id: "alerts",
      name: "Alerts",
      disabled: !config.version,
    },
    {
      id: "logs",
      name: "Logs",
      disabled: !config.version,
    },
    {
      id: "monitoring_dashboard_link",
      name: (
        <span>
          Monitoring&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </span>
      ),
      href: router.monitoring_url,
      disabled: !config.version,
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

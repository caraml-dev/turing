import React from "react";
import { EuiIcon } from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";
import { useConfig } from "../../../../config";

export const RouterDetailsPageNavigation = ({
  router: { config = {}, ...router },
  actions,
  selectedTab,
}) => {
  const { alertConfig } = useConfig();

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
      disabled: !alertConfig.enabled || !config.version,
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
      disabled: !router.monitoring_url || !config.version,
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

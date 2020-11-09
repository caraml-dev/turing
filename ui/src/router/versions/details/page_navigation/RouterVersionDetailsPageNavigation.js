import React from "react";
import { PageNavigation } from "../../../../components/page_navigation/PageNavigation";
import { useMonitoring } from "../../../../hooks/useMonitoring";
import { EuiIcon } from "@elastic/eui";

export const RouterVersionDetailsPageNavigation = ({
  version,
  actions,
  ...props
}) => {
  const [getMonitoringDashboardUrl] = useMonitoring();

  const tabs = [
    {
      id: "details",
      name: "Configuration"
    },
    {
      id: "monitoring_dashboard_link",
      name: (
        <span>
          Monitoring&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </span>
      ),
      href: getMonitoringDashboardUrl(
        version.router.environment_name,
        version.router.name,
        version.version
      )
    }
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

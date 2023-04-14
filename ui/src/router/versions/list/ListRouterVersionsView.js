import React, { useEffect } from "react";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate } from "react-router-dom";
import { useTuringApi } from "../../../hooks/useTuringApi";
import { EuiPanel } from "@elastic/eui";
import { ConfigSection } from "../../../components/config_section";
import { ListRouterVersionsTable } from "./ListRouterVersionsTable";
import "./ListRouterVersionsView.scss";
import { RouterVersionActions } from "../components/RouterVersionActions";

export const ListRouterVersionsView = ({ router }) => {
  const navigate = useNavigate();
  const deployedVersion = (router.config || {}).version;

  const [versions, fetchVersions] = useTuringApi(
    `/projects/${router.project_id}/routers/${router.id}/versions`,
    {},
    [],
    false
  );

  useEffect(() => {
    fetchVersions();
  }, [router.status, fetchVersions]);

  const onRowClick = (item) => {
    navigate(`../versions/${item.version}/details`);
  };

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: `../`,
      },
      {
        text: router.name,
        href: `./`,
      },
      {
        text: "Versions",
      },
    ]);
  }, [router.name]);

  return (
    <ConfigSection title="Versions">
      <EuiPanel>
        <RouterVersionActions
          router={router}
          onDeploySuccess={() =>
            navigate("./", { state: { refresh: true } })
          }
          onDeleteSuccess={fetchVersions}>
          {(actions) => (
            <ListRouterVersionsTable
              items={versions.data}
              isLoaded={versions.isLoaded}
              error={versions.error}
              deployedVersion={deployedVersion}
              onRowClick={onRowClick}
              renderActions={() =>
                actions.map((action) => ({
                  ...action,
                  type: "icon",
                  name: action.label,
                  description: action.name,
                }))
              }
            />
          )}
        </RouterVersionActions>
      </EuiPanel>
    </ConfigSection>
  );
};

import React, { useEffect } from "react";
import {
  EuiButton,
  EuiPageTemplate,
  EuiPanel
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useTuringApi } from "../../hooks/useTuringApi";
import ListRoutersTable from "./ListRoutersTable";

export const ListRoutersView = ({ projectId, ...props }) => {
  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/routers`,
    {},
    []
  );

  useEffect(() => {
    replaceBreadcrumbs([{ text: "Routers" }]);
  }, []);

  const onRowClick = (item) => props.navigate(`./${item.id}/details`);

  return (
    <EuiPageTemplate restrictWidth="95%">
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"graphApp"}
        pageTitle={"Routers"}
        rightSideItems={[
          <EuiButton onClick={() => props.navigate("./create")} fill>
            Create Router
          </EuiButton>,
        ]}
      />
      <EuiPageTemplate.Section restrictWidth="90%" color={"transparent"}>
        <EuiPanel>
          <ListRoutersTable
            isLoaded={isLoaded}
            items={data}
            error={error}
            onRowClick={onRowClick}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

import React, { useEffect } from "react";
import {
  EuiButton,
  EuiPageTemplate,
  EuiPanel, EuiSpacer
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useTuringApi } from "../../hooks/useTuringApi";
import ListRoutersTable from "./ListRoutersTable";
import { PageTitle } from "../../components/page/PageTitle";
import { useConfig } from "../../config";

export const ListRoutersView = ({ projectId, ...props }) => {
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

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
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Routers" />}
        rightSideItems={[
          <EuiButton onClick={() => props.navigate("./create")} fill>
            Create Router
          </EuiButton>,
        ]}
      />

      <EuiSpacer size="m" />
      <EuiPageTemplate.Section  color={"transparent"}>
        <EuiPanel>
          <ListRoutersTable
            isLoaded={isLoaded}
            items={data}
            error={error}
            onRowClick={onRowClick}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

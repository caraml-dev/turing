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
import { useNavigate, useParams } from "react-router-dom";

export const ListRoutersView = () => {
  const { projectId } = useParams();
  const navigate = useNavigate();
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

  const onRowClick = (item) => navigate(`./${item.id}/details`);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Routers" />}
        rightSideItems={[
          <EuiButton onClick={() => navigate("./create")} fill>
            Create Router
          </EuiButton>,
        ]}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
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

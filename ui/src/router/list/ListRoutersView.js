import React, { useEffect } from "react";
import {
  EuiButton,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { Link } from "@reach/router";
import { useTuringApi } from "../../hooks/useTuringApi";
import ListRoutersTable from "./ListRoutersTable";
import { PageTitle } from "../../components/page/PageTitle";

export const ListRoutersView = ({ projectId, ...props }) => {
  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/routers`,
    {},
    []
  );

  useEffect(() => {
    replaceBreadcrumbs([{ text: "Routers" }]);
  }, []);

  const onRowClick = item => props.navigate(`./${item.id}/details`);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Routers" />
          </EuiPageHeaderSection>
          <EuiPageHeaderSection>
            <Link to={"./create"}>
              <EuiButton fill>Create Router</EuiButton>
            </Link>
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContent>
          <ListRoutersTable
            isLoaded={isLoaded}
            items={data}
            error={error}
            onRowClick={onRowClick}
          />
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

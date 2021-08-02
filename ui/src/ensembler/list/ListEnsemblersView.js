import React, { useEffect, useState } from "react";
import { useTuringApi } from "../../hooks/useTuringApi";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import {
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection
} from "@elastic/eui";
import { PageTitle } from "../../components/page/PageTitle";
import { ListEnsemblersTable } from "./ListEnsemblersTable";
import { appConfig } from "../../config";

const { defaultPageSize } = appConfig.pagination;

export const ListEnsemblersView = ({ projectId, ...props }) => {
  const [pagination, setPagination] = useState({ page_size: defaultPageSize });
  const [filter, setFilter] = useState({});
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });

  const onQueryChange = query => {
    setFilter(filter => {
      const typeClause = query.getSimpleFieldClause("type");

      if (!!typeClause) {
        filter["type"] = typeClause.value;
      } else {
        delete filter["type"];
      }
      return { ...filter };
    });
  };

  const onPaginationChange = page => {
    setPagination({
      page: page.index + 1,
      page_size: page.size
    });
  };

  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/ensemblers`,
    {
      query: {
        ...filter,
        ...pagination
      }
    },
    { results: [], paging: { total: 0 } }
  );

  useEffect(() => {
    if (isLoaded && !error) {
      setResults({
        items: data.results,
        totalItemCount: data.paging.total
      });
    }
  }, [data, isLoaded, error]);

  useEffect(() => {
    replaceBreadcrumbs([{ text: "Ensemblers" }]);
  }, []);

  const onRowClick = item => {};
  // props.navigate(`./${item.id}/details`);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Ensemblers" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContent>
          <ListEnsemblersTable
            {...results}
            isLoaded={isLoaded}
            error={error}
            defaultPageSize={defaultPageSize}
            onQueryChange={onQueryChange}
            onPaginationChange={onPaginationChange}
            onRowClick={onRowClick}
          />
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

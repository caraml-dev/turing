import { useTuringApi } from "../../hooks/useTuringApi";
import {
  EuiButton,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection
} from "@elastic/eui";
import { PageTitle } from "../../components/page/PageTitle";
import React, { useEffect, useState } from "react";
import { ListEnsemblingJobsTable } from "./ListEnsemblingJobsTable";
import { EnsemblersContextContextProvider } from "../../providers/ensemblers/context";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { appConfig } from "../../config";

const { defaultPageSize } = appConfig.pagination;

export const ListEnsemblingJobsView = ({ projectId }) => {
  const [pagination, setPagination] = useState({ page_size: defaultPageSize });
  const [filter, setFilter] = useState({});
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });

  const onQueryChange = query => {
    setFilter(filter => {
      const statusClause = query.getOrFieldClause("status");

      if (!!statusClause) {
        filter["status"] = statusClause.value;
      } else {
        delete filter["status"];
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
    `/projects/${projectId}/jobs`,
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
    replaceBreadcrumbs([{ text: "Jobs" }]);
  }, []);

  const onRowClick = item => {};
  // props.navigate(`./${item.id}/details`);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Ensembling Jobs" />
          </EuiPageHeaderSection>
          <EuiPageHeaderSection>
            <EuiButton fill disabled href={"jobs/create"}>
              Submit Job
            </EuiButton>
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContent>
          <EnsemblersContextContextProvider projectId={projectId}>
            <ListEnsemblingJobsTable
              {...results}
              isLoaded={isLoaded}
              error={error}
              defaultPageSize={defaultPageSize}
              onQueryChange={onQueryChange}
              onPaginationChange={onPaginationChange}
              onRowClick={onRowClick}
            />
          </EnsemblersContextContextProvider>
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

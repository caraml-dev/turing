import { useTuringApi } from "../../hooks/useTuringApi";
import {
  EuiButton,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection,
} from "@elastic/eui";
import { PageTitle } from "../../components/page/PageTitle";
import React, { useEffect, useMemo, useState } from "react";
import { ListEnsemblingJobsTable } from "./ListEnsemblingJobsTable";
import { EnsemblersContextContextProvider } from "../../providers/ensemblers/context";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { appConfig } from "../../config";
import { parse, stringify } from "query-string";

const { defaultPageSize } = appConfig.pagination;

export const ListEnsemblingJobsView = ({ projectId, ...props }) => {
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({ index: 0, size: defaultPageSize });
  const filter = useMemo(
    () => parse(props.location.search),
    [props.location.search]
  );

  const onQueryChange = ({ query }) => {
    const filter = {};
    const ensemblerClause = query.getSimpleFieldClause("ensembler_id");
    if (!!ensemblerClause) {
      filter["ensembler_id"] = ensemblerClause.value;
    }

    const statusClause = query.getOrFieldClause("status");
    if (!!statusClause) {
      filter["status"] = statusClause.value;
    }

    const searchClause = query.ast.getTermClauses();
    if (!!searchClause) {
      filter["search"] = searchClause.map((c) => c.value).join(" ");
    }

    props.navigate(`${props.location.pathname}?${stringify(filter)}`);
  };

  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/jobs`,
    {
      query: {
        ...filter,
        ...{
          page: page.index + 1,
          page_size: page.size,
        },
      },
    },
    { results: [], paging: { total: 0 } }
  );

  useEffect(() => {
    if (isLoaded && !error) {
      setResults({
        items: data.results,
        totalItemCount: data.paging.total,
      });
    }
  }, [data, isLoaded, error]);

  useEffect(() => {
    replaceBreadcrumbs([{ text: "Jobs" }]);
  }, []);

  const onRowClick = (item) => props.navigate(`./${item.id}/details`);

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
              page={page}
              filter={filter}
              onQueryChange={onQueryChange}
              onPaginationChange={setPage}
              onRowClick={onRowClick}
              {...props}
            />
          </EnsemblersContextContextProvider>
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

import React, { useEffect, useMemo, useState } from "react";
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
import { parse, stringify } from "query-string";

const { defaultPageSize } = appConfig.pagination;

export const ListEnsemblersView = ({ projectId, ...props }) => {
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({ index: 0, size: defaultPageSize });
  const filter = useMemo(() => parse(props.location.search), [
    props.location.search
  ]);

  const onQueryChange = ({ query }) => {
    const filter = {};
    const typeClause = query.getSimpleFieldClause("type");
    if (!!typeClause) {
      filter["type"] = typeClause.value;
    }

    const searchClause = query.ast.getTermClauses();
    if (!!searchClause) {
      filter["search"] = searchClause.map(c => c.value).join(" ");
    }

    props.navigate(`${props.location.pathname}?${stringify(filter)}`);
  };

  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectId}/ensemblers`,
    {
      query: {
        ...filter,
        ...{
          page: page.index + 1,
          page_size: page.size
        }
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
            page={page}
            filter={filter}
            onQueryChange={onQueryChange}
            onPaginationChange={setPage}
            onRowClick={onRowClick}
            {...props}
          />
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

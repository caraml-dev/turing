import { useTuringApi } from "../../hooks/useTuringApi";
import {
  EuiButton,
  EuiPageTemplate,
  EuiPanel, EuiSpacer
} from "@elastic/eui";
import { useLocation, useNavigate } from "react-router-dom";
import { PageTitle } from "../../components/page/PageTitle";
import React, { useEffect, useMemo, useState } from "react";
import { ListEnsemblingJobsTable } from "./ListEnsemblingJobsTable";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useConfig } from "../../config";
import { parse, stringify } from "query-string";

export const ListEnsemblingJobsView = ({ projectId }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const {
    appConfig: {
      pagination: { defaultPageSize },
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({ index: 0, size: defaultPageSize });
  const filter = useMemo(
    () => parse(location.search),
    [location.search]
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

    navigate(`${location.pathname}?${stringify(filter)}`);
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

  const onRowClick = (item) => navigate(`./${item.id}/details`);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Ensembling Jobs" />}
        rightSideItems={[
          <EuiButton fill disabled href={"jobs/create"}>
            Submit Job
          </EuiButton>,
        ]}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <ListEnsemblingJobsTable
            {...results}
            isLoaded={isLoaded}
            error={error}
            page={page}
            filter={filter}
            onQueryChange={onQueryChange}
            onPaginationChange={setPage}
            onRowClick={onRowClick}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

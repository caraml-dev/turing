import { useTuringApi } from "../../hooks/useTuringApi";
import {
  EuiButton,
  EuiPageTemplate,
  EuiPanel
} from "@elastic/eui";
// import { PageTitle } from "../../components/page/PageTitle";
import React, { useEffect, useMemo, useState } from "react";
import { ListEnsemblingJobsTable } from "./ListEnsemblingJobsTable";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useConfig } from "../../config";
import { parse, stringify } from "query-string";

export const ListEnsemblingJobsView = (props) => {
  const projectId = props.projectId;
  const {
    appConfig: {
      pagination: { defaultPageSize },
    },
  } = useConfig();
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
    <EuiPageTemplate restrictWidth="95%">
      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"graphApp"}
        pageTitle={"Ensembling Jobs"}
        rightSideItems={[
          <EuiButton fill disabled href={"jobs/create"}>
            Submit Job
          </EuiButton>,
        ]}
      />
      <EuiPageTemplate.Section restrictWidth="90%" color={"transparent"}>
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
            {...props}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

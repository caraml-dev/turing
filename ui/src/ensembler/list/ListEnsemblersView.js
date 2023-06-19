import React, { useEffect, useMemo, useState } from "react";
import { useTuringApi } from "../../hooks/useTuringApi";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { EuiPageTemplate, EuiPanel, EuiSpacer } from "@elastic/eui";
import { PageTitle } from "../../components/page/PageTitle";
import { ListEnsemblersTable } from "./ListEnsemblersTable";
import { useConfig } from "../../config";
import { parse, stringify } from "query-string";
import { useLocation, useNavigate, useParams } from "react-router-dom";

export const ListEnsemblersView = () => {
  const { projectId } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
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
    const typeClause = query.getSimpleFieldClause("type");
    if (!!typeClause) {
      filter["type"] = typeClause.value;
    }

    const searchClause = query.ast.getTermClauses();
    if (!!searchClause) {
      filter["search"] = searchClause.map((c) => c.value).join(" ");
    }

    navigate(`${location.pathname}?${stringify(filter)}`);
  };

  const [{ data, isLoaded, error }, fetchEnsemblerData] = useTuringApi(
    `/projects/${projectId}/ensemblers`,
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
    replaceBreadcrumbs([{ text: "Ensemblers" }]);
  }, []);

  const onDeleteSuccess = () => {
    fetchEnsemblerData()
  }

  const onRowClick = (_item) => { };

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Ensemblers" />}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <ListEnsemblersTable
            {...results}
            isLoaded={isLoaded}
            error={error}
            page={page}
            filter={filter}
            onQueryChange={onQueryChange}
            onPaginationChange={setPage}
            onRowClick={onRowClick}
            onDeleteSuccess={onDeleteSuccess}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

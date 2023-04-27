import React, { Fragment, useEffect, useState } from "react";
import {
  EuiCallOut,
  EuiBasicTable,
  EuiText,
} from "@elastic/eui";
import { DeploymentStatusHealth } from "../../../components/status_health/DeploymentStatusHealth";
import { Status } from "../../../services/status/Status";
import { useConfig } from "../../../config";
import { useTuringApi } from "../../../hooks/useTuringApi";

export const ListRouterVersionsForEnsemblerTable = ({
  projectID,
  ensemblerID
}) => {

  const [results, setResults] = useState({ items: [], totalItemCount: 0 });

  const {
    appConfig: {
      tables: { defaultTextSize },
    },
  } = useConfig();

  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectID}/routers-version-ensembler/${ensemblerID}?status=failed&status=undeployed`,
    []
  )

  useEffect(() => {
    if (isLoaded && !error) {
      setResults({
        items: data,
        totalItemCount: data.length,
      });
    }
  }, [data, isLoaded, error]);


  const columns = [
    {
      field: "id",
      name: "Id",
      width: "96px",
      render: (id, item) => (
        <EuiText size={defaultTextSize}>
          {id}
        </EuiText>
      ),
    },
    {
      field: "name",
      name: "Name",
      truncateText: true,
      render: (id, item) => (
        <span className="eui-textTruncate" title={item.router.name}>
          {item.router.name} - version {item.version}
        </span>
      ),
    },
    {
      field: "status",
      name: "Status",
      width: "150px",
      render: (status) => (
        <DeploymentStatusHealth status={Status.fromValue(status)} />
      ),
    },
  ];

  return error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <Fragment>
      {results.totalItemCount > 0 && ( 
        <div>
          <p>Deleting this Ensembler will also delete {results.totalItemCount} <b>Failed</b> or <b>Undeployed</b> Router Versions that use this Ensembler </p>
          <EuiBasicTable
            items={results.items}
            loading={!isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
          />
        </div>
      )}
    </Fragment>
  );
};

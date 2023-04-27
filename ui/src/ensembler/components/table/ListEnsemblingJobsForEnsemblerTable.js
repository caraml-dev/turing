import React, { Fragment, useEffect, useState } from "react";
import {
  EuiCallOut,
  EuiBasicTable,
  EuiText,
} from "@elastic/eui";
import { DeploymentStatusHealth } from "../../../components/status_health/DeploymentStatusHealth";
import { JobStatus } from "../../../services/job/JobStatus";
import { useConfig } from "../../../config";
import { useTuringApi } from "../../../hooks/useTuringApi";


export const ListEnsemblingJobsForEnsemblerTable = ({
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
    `/projects/${projectID}/jobs?ensembler_id=${ensemblerID}&status=failed_submission&status=failed_building&status=failed&status=completed`,
    { results: [], paging: { total: 0 } }
  )

  useEffect(() => {
    if (isLoaded && !error) {
      setResults({
        items: data.results,
        totalItemCount: data.paging.total,
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
      render: (name) => (
        <span className="eui-textTruncate" title={name}>
          {name}
        </span>
      ),
    },
    {
      field: "status",
      name: "Status",
      width: "120px",
      render: (status) => (
        <DeploymentStatusHealth status={JobStatus.fromValue(status)} />
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
          <p>Deleting this Ensembler will also delete {results.totalItemCount} <b>Failed</b> or <b>Completed</b> Ensembling Jobs that use this Ensembler </p>
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

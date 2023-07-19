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
  ensemblerID,
  canDeleteEnsembler,
  setEnsemblerUsedByActiveEnsemblingJob
}) => {
  const [results, setResults] = useState({ inactiveItems: [], activeItems:[], totalInactiveCount: 0, totalActiveCount:0 });

  const {
    appConfig: {
      tables: { defaultTextSize },
    },
  } = useConfig();

  const [{ data, isLoaded, error }] = useTuringApi(
    `/projects/${projectID}/jobs?ensembler_id=${ensemblerID}`,
    []
  )

  useEffect(() => {
    if (isLoaded && !error) {
      let inactiveItems = data.results.filter((item) => ["failed", "failed_submission", "failed_building", "completed"].includes(item.status));
      let activeItems = data.results.filter((item) => !["failed", "failed_submission", "failed_building", "completed"].includes(item.status));
      setResults({
        inactiveItems: inactiveItems ,
        activeItems: activeItems ,
        totalInactiveCount: inactiveItems.length,
        totalActiveCount : activeItems.length
      });
    }
  }, [data, isLoaded, error]);

  useEffect(() => {
    if (results.activeItems.length > 0){
      setEnsemblerUsedByActiveEnsemblingJob(true)
    } else {
      setEnsemblerUsedByActiveEnsemblingJob(false)
    }
  }, [results, setEnsemblerUsedByActiveEnsemblingJob])

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
      render: (_, item) => (
        <span className="eui-textTruncate" title={item.name}>
          <a href={`./jobs/${item.id}/details`} target="_blank" rel="noreferrer">{item.name}</a>
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
      {canDeleteEnsembler ? ( results.totalInactiveCount > 0 && (
        <div>
          <br/>
          Deleting this Ensembler will also delete {results.totalInactiveCount} <b>Failed</b> or <b>Completed</b>
          &nbsp;Ensembling Job{results.totalInactiveCount > 1 && "s"} that use{results.totalInactiveCount === 1 && "s"} this Ensembler:
          <EuiBasicTable
            items={results.inactiveItems}
            loading={!isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
          />
        </div>
      )) : ( results.totalActiveCount > 0 && (
        <div>
          <br/>
          This Ensembler is used by {results.totalActiveCount} <b>Active Ensembling Job{results.totalActiveCount > 1 && "s"}</b>.
          If any ensembling jobs are in the terminating state, please wait until they are complete to delete this ensembler.
          <EuiBasicTable
            items={results.activeItems}
            loading={!isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
          />
          If you wish to delete this ensembler, please <b>terminate</b> {results.totalActiveCount > 1 ? "these ensembling jobs" : "this ensembling job"}.
        </div>
      ))}
    </Fragment>
  );
};

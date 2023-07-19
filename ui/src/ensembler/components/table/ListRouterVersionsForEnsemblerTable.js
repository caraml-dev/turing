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
  ensemblerID,
  canDeleteEnsembler,
  setEnsemblerUsedByActiveRouterVersion,
  setEnsemblerUsedByCurrentRouterVersion,
  ensemblerUsedByCurrentRouterVersion
}) => {
  const [results, setResults] = useState({ inactiveItems: [], activeItems:[], totalInactiveCount: 0, totalActiveCount:0 });
  const [filteredActiveRouterVersions, setFilteredActiveRouterVersions] = useState([])

  const {
    appConfig: {
      tables: { defaultTextSize },
    },
  } = useConfig();

  const [ allRouterVersion ] = useTuringApi(
    `/projects/${projectID}/router-versions?ensembler_id=${ensemblerID}`,
    []
  )

  const [ currentRouterVersion ] = useTuringApi(
    `/projects/${projectID}/router-versions?ensembler_id=${ensemblerID}&is_current=true`,
    []
  )

  useEffect(() => {
    if (allRouterVersion.isLoaded && !allRouterVersion.error) {
      let inactiveItems = allRouterVersion.data.filter((item) => item.status === 'failed' || item.status==='undeployed');
      let activeItems = allRouterVersion.data.filter((item) => item.status !== 'failed' && item.status!=='undeployed')
      setResults({
        inactiveItems: inactiveItems ,
        activeItems: activeItems ,
        totalInactiveCount: inactiveItems.length,
        totalActiveCount : activeItems.length
      });
    }
  }, [allRouterVersion]);

  useEffect(() => {
    if (results.activeItems.length > 0){
      setEnsemblerUsedByActiveRouterVersion(true);
    }

    if (currentRouterVersion.isLoaded && !currentRouterVersion.error) {
      if (currentRouterVersion.data.length > 0){
        setEnsemblerUsedByCurrentRouterVersion(true);
        setFilteredActiveRouterVersions(results.activeItems.filter(item => currentRouterVersion.data.every((e => e.id !== item.id))));
      } else {
        setEnsemblerUsedByCurrentRouterVersion(false)
      }
    }
  }, [results, currentRouterVersion, setEnsemblerUsedByActiveRouterVersion, setEnsemblerUsedByCurrentRouterVersion])

  const columns = [
    {
      field: "id",
      name: "Id",
      width: "96px",
      render: (id, _) => (
        <EuiText size={defaultTextSize}>
          {id}
        </EuiText>
      ),
    },
    {
      field: "router version",
      name: "Version",
      width: "96px",
      render: (_, item) => (
        <EuiText size={defaultTextSize}>
          Version {item.version}
        </EuiText>
      ),
    },
    {
      field: "name",
      name: "Router Name",
      truncateText: true,
      render: (_, item) => (
        <span className="eui-textTruncate" title={item.router.name}>
          <a href={`./routers/${item.router.id}/history`} target="_blank" rel="noreferrer">{item.router.name}</a>
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

  return allRouterVersion.error || currentRouterVersion.error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      {allRouterVersion.error && <p>{allRouterVersion.error.message}</p>}
      {currentRouterVersion.error && <p>{currentRouterVersion.error.message}</p>}
    </EuiCallOut>
  ) : (
    <Fragment>
      {canDeleteEnsembler && results.totalInactiveCount > 0 && (
        <div>
          <br/>
          Deleting this Ensembler will also delete {results.totalInactiveCount} <b>Failed</b> or <b>Undeployed</b>
          &nbsp;Router Version{results.totalInactiveCount > 1 && "s"} that
          use{results.totalInactiveCount === 1 && "s"} this Ensembler:
          <EuiBasicTable
            items={results.inactiveItems}
            loading={!allRouterVersion.isLoaded && !currentRouterVersion.isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
          />
        </div>
      )}
      {ensemblerUsedByCurrentRouterVersion && currentRouterVersion.data.length > 0 && (
        <div>
          This Ensembler is used by {currentRouterVersion.data.length}
          &nbsp;<b>Router{currentRouterVersion.data.length > 1 && "s"}</b>
          &nbsp;in {currentRouterVersion.data.length > 1 ? "their" : "its"} <b>current</b> configuration:
          <EuiBasicTable
            items={currentRouterVersion.data}
            loading={!allRouterVersion.isLoaded && !currentRouterVersion.isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
          />
          If you wish to delete this ensembler, please <b>deploy</b> another router version
          for {currentRouterVersion.data.length > 1 ? "these routers" : "this router"}.
        </div>
      )}
      {filteredActiveRouterVersions.length > 0 && (
        <div>
          <br/>
          This Ensembler is {ensemblerUsedByCurrentRouterVersion && currentRouterVersion.data.length > 0 && "also"}
          &nbsp;used by {filteredActiveRouterVersions.length}
          &nbsp;<b>Active Router Version{filteredActiveRouterVersions.length > 1 && "s"}</b>:
          <EuiBasicTable
            items={filteredActiveRouterVersions}
            loading={!allRouterVersion.isLoaded && !currentRouterVersion.isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
          />
          If you wish to delete this ensembler, please <b>undeploy</b>
          &nbsp;{filteredActiveRouterVersions.length > 1 ? "these router versions" : "this router version"}.
        </div>
      )}
    </Fragment>
  );
};

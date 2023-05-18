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
  setCanDeleteEnsembler,
  canDeleteEnsembler,
  setEnsemblerUsedByCurrentRouterVersion,
  ensemblerUsedByCurrentRouterVersion
}) => {
  const [results, setResults] = useState({ inactiveItems: [], activeItems:[], totalInactiveCount: 0, totalActiveCount:0 });

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
      setCanDeleteEnsembler(false)
    } else {
      if (currentRouterVersion.isLoaded && !currentRouterVersion.error) {
        if (currentRouterVersion.data.length > 0){
          setEnsemblerUsedByCurrentRouterVersion(true)
          setCanDeleteEnsembler(false)
        } else {
          setEnsemblerUsedByCurrentRouterVersion(false)
          setCanDeleteEnsembler(true)
        }
      }
    }
  }, [results, currentRouterVersion, setCanDeleteEnsembler, setEnsemblerUsedByCurrentRouterVersion])

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
      field: "router version",
      name: "Version",
      width: "96px",
      render: (id, item) => (
        <EuiText size={defaultTextSize}>
          Version {item.version}
        </EuiText>
      ),
    },
    {
      field: "name",
      name: "Router Name",
      truncateText: true,
      render: (id, item) => (
        <span className="eui-textTruncate" title={item.router.name}>
          {item.router.name}
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

  const cellProps = (item) =>
    ({
      style: { cursor: "pointer" },
      onClick: () => window.open(`./routers/${item.id}/history`, '_blank'),
    });

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
      {ensemblerUsedByCurrentRouterVersion ? (
        currentRouterVersion.data.length > 0 && (
        <div>
          <br/>
          <p>The router version with the related ensembler is being used by {currentRouterVersion.data.length} <b>Routers</b></p>
          <EuiBasicTable
            items={results.inactiveItems}
            loading={!allRouterVersion.isLoaded && !currentRouterVersion.isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
            cellProps={cellProps}
          />
        </div>
      )) : canDeleteEnsembler ? ( results.totalInactiveCount > 0 && (
        <div>
          <br/>
          <p>Deleting this Ensembler will also delete {results.totalInactiveCount} <b>Failed</b> or <b>Undeployed</b> Router Versions that use this Ensembler </p>
          <EuiBasicTable
            items={results.inactiveItems}
            loading={!allRouterVersion.isLoaded && !currentRouterVersion.isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
            cellProps={cellProps}
          />
        </div>
      )) : ( results.totalActiveCount > 0 && (
        <div>
          <br/>
          <p>This Ensembler is being used by {results.totalActiveCount} <b>Active Router Versions</b></p>
          <EuiBasicTable
            items={results.activeItems}
            loading={!allRouterVersion.isLoaded && !currentRouterVersion.isLoaded}
            columns={columns}
            responsive={true}
            tableLayout="auto"
            cellProps={cellProps}
          />
        </div>
      ))}
    </Fragment>
  );
};

import React, { useContext, Fragment, useState } from "react";
import {
  EuiBadge,
  EuiCallOut,
  EuiHealth,
  EuiIcon,
  EuiBasicTable,
  EuiLink,
  EuiSearchBar,
  EuiSpacer,
  EuiText,
  EuiToolTip
} from "@elastic/eui";
import { appConfig } from "../../config";
import moment from "moment";
import { DeploymentStatusHealth } from "../../components/status_health/DeploymentStatusHealth";
import { JobStatus } from "../../services/job_status/JobStatus";
import EnsemblersContext from "../../providers/ensemblers/context";

const { defaultTextSize, defaultIconSize, dateFormat } = appConfig.tables;

export const ListEnsemblingJobsTable = ({
  items,
  totalItemCount,
  isLoaded,
  error,
  page,
  onQueryChange,
  onPaginationChange,
  onRowClick
}) => {
  const ensemblers = useContext(EnsemblersContext);
  const [searchQuery, setSearchQuery] = useState(EuiSearchBar.Query.MATCH_ALL);

  const onSearchChange = ({ queryText }) => {
    const query = EuiSearchBar.Query.parse(queryText);
    setSearchQuery(query);
    onQueryChange(query);
  };

  const onTableChange = ({ page = {} }) => onPaginationChange(page);

  const columns = [
    {
      field: "id",
      name: "Id",
      width: "72px",
      render: (id, item) => (
        <EuiText size={defaultTextSize}>
          {id}
          {moment().diff(item.created_at, "hours") <= 1 && (
            <Fragment>
              &nbsp;
              <EuiBadge color="secondary">New</EuiBadge>
            </Fragment>
          )}
        </EuiText>
      )
    },
    {
      field: "name",
      name: "Name",
      width: "30%",
      render: name => (
        <span className="eui-textTruncate" title={name}>
          {name}
        </span>
      )
    },
    {
      field: "ensembler_id",
      name: "Ensembler",
      width: "20%",
      render: id =>
        !!ensemblers[id] ? (
          <EuiLink href={`./ensemblers/${id}`}>
            <EuiIcon type={"aggregate"} size={defaultIconSize} />
            {ensemblers[id].name}
          </EuiLink>
        ) : (
          "[Removed]"
        )
    },
    {
      field: "status",
      name: "Status",
      width: "20%",
      render: status => (
        <DeploymentStatusHealth status={JobStatus.fromValue(status)} />
      )
    },
    {
      field: "created_at",
      name: "Created",
      sortable: true,
      width: "20%",
      render: date => (
        <EuiToolTip
          position="top"
          content={moment(date, dateFormat).toLocaleString()}>
          <EuiText size={defaultTextSize}>
            {moment(date, dateFormat).fromNow()}
          </EuiText>
        </EuiToolTip>
      )
    },
    {
      field: "updated_at",
      name: "Updated",
      width: "20%",
      render: date => (
        <EuiToolTip
          position="top"
          content={moment(date, dateFormat).toLocaleString()}>
          <EuiText size={defaultTextSize}>
            {moment(date, dateFormat).fromNow()}
          </EuiText>
        </EuiToolTip>
      )
    }
  ];

  const pagination = {
    pageIndex: page.index,
    pageSize: page.size,
    totalItemCount
  };

  const search = {
    query: searchQuery,
    onChange: onSearchChange,
    box: {
      incremental: true
    },
    filters: [
      {
        type: "field_value_selection",
        field: "status",
        name: "Status",
        multiSelect: "or",
        options: JobStatus.values.map(status => ({
          value: status.toString(),
          view: <EuiHealth color={status.color}>{status.toString()}</EuiHealth>
        }))
      }
    ]
  };

  const cellProps = item =>
    onRowClick
      ? {
          style: { cursor: "pointer" },
          onClick: () => onRowClick(item)
        }
      : undefined;

  return error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <Fragment>
      <EuiSearchBar {...search} />
      <EuiSpacer size="l" />
      <EuiBasicTable
        items={items}
        loading={!isLoaded}
        columns={columns}
        cellProps={cellProps}
        pagination={pagination}
        onChange={onTableChange}
      />
    </Fragment>
  );
};

import React, { Fragment, useContext, useMemo } from "react";
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
} from "@elastic/eui";
import { useNavigate } from "react-router-dom";
import { useConfig } from "../../config";
import moment from "moment";
import { DeploymentStatusHealth } from "../../components/status_health/DeploymentStatusHealth";
import { JobStatus } from "../../services/job/JobStatus";
import EnsemblersContext from "../../providers/ensemblers/context";
import { DateFromNow } from "@gojek/mlp-ui";

export const ListEnsemblingJobsTable = ({
  items,
  totalItemCount,
  isLoaded,
  error,
  page,
  filter,
  onQueryChange,
  onPaginationChange,
  onRowClick,
}) => {
  const navigate = useNavigate();
  const {
    appConfig: {
      tables: { defaultTextSize, defaultIconSize },
    },
  } = useConfig();
  const { ensemblers } = useContext(EnsemblersContext);

  const searchQuery = useMemo(() => {
    const parts = [];
    if (!!filter.search) {
      parts.push(filter.search);
    }
    if (!!filter.ensembler_id) {
      parts.push(`ensembler_id:${filter.ensembler_id}`);
    }
    if (!!filter.status) {
      const statuses = Array.isArray(filter.status)
        ? filter.status
        : [filter.status];
      parts.push(`status:(${statuses.join(" or ")})`);
    }

    return parts.join(" ");
  }, [filter]);

  const onTableChange = ({ page = {} }) => onPaginationChange(page);

  const columns = [
    {
      field: "id",
      name: "Id",
      width: "96px",
      render: (id, item) => (
        <EuiText size={defaultTextSize}>
          {id}
          {moment().diff(item.created_at, "hours") <= 1 && (
            <Fragment>
              &nbsp;
              <EuiBadge color="success">New</EuiBadge>
            </Fragment>
          )}
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
      field: "ensembler_id",
      name: "Ensembler",
      render: (id) =>
        !!ensemblers[id] ? (
          <EuiLink
            onClick={(e) => {
              e.stopPropagation();
              navigate(`./?ensembler_id=${id}`);
            }}>
            <EuiIcon type={"aggregate"} size={defaultIconSize} />
            {ensemblers[id].name}
          </EuiLink>
        ) : (
          "[Removed]"
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
    {
      field: "created_at",
      name: "Created",
      sortable: true,
      render: (date) => <DateFromNow date={date} size={defaultTextSize} />,
    },
    {
      field: "updated_at",
      name: "Updated",
      render: (date) => <DateFromNow date={date} size={defaultTextSize} />,
    },
    {
      name: "Actions",
      actions: [
        {
          name: "Monitoring",
          description: "Open Monitoring Dashboard",
          icon: "visLine",
          type: "icon",
          isPrimary: true,
          onClick: ({ monitoring_url }) => {
            window.open(monitoring_url, "_blank");
          },
          enabled: ({ monitoring_url }) => !!monitoring_url,
        },
      ],
    },
  ];

  const pagination = {
    pageIndex: page.index,
    pageSize: page.size,
    totalItemCount,
  };

  const search = {
    query: searchQuery,
    onChange: onQueryChange,
    box: {
      incremental: false,
    },
    filters: [
      {
        type: "field_value_selection",
        field: "status",
        name: "Status",
        multiSelect: "or",
        options: JobStatus.values.map((status) => ({
          value: status.toString(),
          view: <EuiHealth color={status.color}>{status.toString()}</EuiHealth>,
        })),
      },
      {
        type: "field_value_selection",
        field: "ensembler_id",
        name: "Ensembler",
        multiSelect: false,
        options: Object.values(ensemblers).map((ensembler) => ({
          value: ensembler.id,
          view: ensembler.name,
        })),
      },
    ],
  };

  const cellProps = (item) =>
    onRowClick
      ? {
        style: { cursor: "pointer" },
        onClick: () => onRowClick(item),
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
        hasActions={true}
        responsive={true}
        tableLayout="auto"
      />
    </Fragment>
  );
};

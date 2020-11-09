import React from "react";
import {
  EuiBadge,
  EuiCallOut,
  EuiInMemoryTable,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign
} from "@elastic/eui";
import { DateFromNow } from "@gojek/mlp-ui";
import { DeploymentStatusHealth } from "../../components/status_health/DeploymentStatusHealth";
import { FormLabelWithToolTip } from "../../../components/form/label_with_tooltip/FormLabelWithToolTip";

const moment = require("moment");

const defaultTextSize = "s";
const defaultIconSize = "s";

export const ListRouterVersionsTable = ({
  items,
  isLoaded,
  error,
  deployedVersion,
  renderActions,
  onRowClick
}) => {
  const cellProps = item =>
    onRowClick
      ? {
          style: { cursor: "pointer" },
          onClick: () => onRowClick(item)
        }
      : undefined;

  const columns = [
    {
      field: "version",
      name: "Version",
      mobileOptions: {
        enlarge: true,
        fullWidth: true
      },
      sortable: true,
      width: "25%",
      render: (version, item) => (
        <EuiText size={defaultTextSize}>
          {version}&nbsp;&nbsp;
          {moment().diff(item.created_at, "hours") <= 1 && (
            <EuiBadge color="secondary">New</EuiBadge>
          )}
          {version === deployedVersion && (
            <EuiBadge color="default">Current</EuiBadge>
          )}
        </EuiText>
      )
    },
    {
      field: "status",
      name: "Status",
      width: "20%",
      render: status => <DeploymentStatusHealth status={status} />
    },
    {
      field: "created_at",
      name: "Created",
      width: "20%",
      render: date => <DateFromNow date={date} size={defaultTextSize} />
    },
    {
      field: "updated_at",
      name: "Updated",
      width: "20%",
      render: date => <DateFromNow date={date} size={defaultTextSize} />
    },
    {
      name: (
        <FormLabelWithToolTip
          label="Actions"
          size={defaultIconSize}
          content="Router version actions"
        />
      ),
      width: "15%",
      align: "right",
      hasActions: true,
      actions: renderActions()
    }
  ];

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <EuiInMemoryTable
      items={items}
      columns={columns}
      cellProps={cellProps}
      itemId="version"
      pagination={true}
      sorting={{ sort: { field: "Version", direction: "desc" } }}
    />
  );
};

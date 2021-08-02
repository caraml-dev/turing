import React, { useState, useEffect } from "react";
import {
  EuiBadge,
  EuiButtonEmpty,
  EuiCallOut,
  EuiInMemoryTable,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign,
  EuiToolTip,
  EuiSearchBar
} from "@elastic/eui";
import { useMonitoring } from "../../hooks/useMonitoring";
import { RouterEndpoint } from "../components/router_endpoint/RouterEndpoint";
import { FormLabelWithToolTip } from "../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { appConfig } from "../../config";
import moment from "moment";
import { DeploymentStatusHealth } from "../../components/status_health/DeploymentStatusHealth";
import { Status } from "../../services/status/Status";

const { defaultTextSize, defaultIconSize, dateFormat } = appConfig.tables;

const ListRoutersTable = ({ items, isLoaded, error, onRowClick }) => {
  const [config, setConfig] = useState({
    environments: []
  });

  const [getMonitoringDashboardUrl] = useMonitoring();

  useEffect(() => {
    if (isLoaded && items.length) {
      let envDict = {};
      items.forEach(item => {
        envDict[item.environment_name] = true;
      });
      setConfig({
        environments: Object.keys(envDict)
      });
    }
  }, [isLoaded, items]);

  const columns = [
    {
      field: "name",
      name: "Name",
      mobileOptions: {
        enlarge: true,
        fullWidth: true
      },
      width: "10%",
      render: (name, item) => (
        <EuiText size={defaultTextSize}>
          {name}&nbsp;
          {moment().diff(item.created_at, "hours") <= 1 && (
            <EuiBadge color="secondary">New</EuiBadge>
          )}
        </EuiText>
      )
    },
    {
      field: "environment_name",
      name: "Environment",
      width: "10%",
      render: environment_name => (
        <EuiText size={defaultTextSize}>{environment_name}</EuiText>
      )
    },
    {
      field: "endpoint",
      name: "Endpoint",
      width: "15%",
      render: endpoint => <RouterEndpoint endpoint={endpoint} />
    },
    {
      field: "status",
      name: "Status",
      width: "10%",
      render: status => (
        <DeploymentStatusHealth status={Status.fromValue(status)} />
      )
    },
    {
      field: "created_at",
      name: "Created",
      sortable: true,
      width: "10%",
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
      width: "10%",
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
      name: (
        <FormLabelWithToolTip
          label="Actions"
          size={defaultIconSize}
          content="Router actions"
        />
      ),
      align: "right",
      mobileOptions: {
        header: true,
        fullWidth: false
      },
      width: "10%",
      render: item => {
        const monitoringLink = item.config
          ? getMonitoringDashboardUrl(item.environment_name, item.name)
          : undefined;

        return (
          <EuiButtonEmpty
            onClick={e => {
              e.stopPropagation();
            }}
            href={monitoringLink}
            isDisabled={!monitoringLink}
            iconType="visLine"
            size="xs"
            target="_blank">
            <EuiText size="xs">Monitoring</EuiText>
          </EuiButtonEmpty>
        );
      }
    }
  ];

  const cellProps = item =>
    onRowClick
      ? {
          style: { cursor: "pointer" },
          onClick: () => onRowClick(item)
        }
      : undefined;

  const onChange = ({ query, error }) => {
    if (error) {
      return error;
    } else {
      return EuiSearchBar.Query.execute(query, items, {
        defaultFields: ["name"]
      });
    }
  };

  const search = {
    onChange: onChange,
    box: {
      incremental: true
    },
    filters: [
      {
        type: "field_value_selection",
        field: "environment_name",
        name: "Environment",
        multiSelect: "or",
        options: config.environments.map(item => ({
          value: item
        }))
      }
    ]
  };

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
      itemId="id"
      search={search}
      sorting={{ sort: { field: "Created", direction: "desc" } }}
    />
  );
};

export default ListRoutersTable;

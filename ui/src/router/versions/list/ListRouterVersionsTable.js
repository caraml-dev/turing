import React, { Fragment, useCallback, useMemo, useState } from "react";
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
import { DefaultItemAction } from "../components/DefaultItemAction";

const defaultTextSize = "s";
const defaultIconSize = "s";

export const ListRouterVersionsTable = ({
  items,
  isLoaded,
  error,
  deployedVersion,
  renderActions,
  onRowClick,
  ...props
}) => {
  const [selection, setSelection] = useState({});

  const isSelected = useCallback(item => !!selection[item.version], [
    selection
  ]);

  const readyForDiff = useMemo(() => Object.entries(selection).length === 2, [
    selection
  ]);

  const selectForComparison = useCallback(
    item => {
      setSelection(items => ({
        ...items,
        [item.version]: item
      }));
    },
    [setSelection]
  );

  const unselectFromComparison = useCallback(
    item => {
      setSelection(items => {
        delete items[item.version];
        return { ...items };
      });
    },
    [setSelection]
  );

  const onShowDiff = useCallback(() => {
    const selected = Object.values(selection);
    if (selected.length === 2) {
      const [leftSelection, rightSelection] = selected;
      props.navigate(
        `./compare/${leftSelection.version}/${rightSelection.version}`,
        {
          state: {
            versionLeft: leftSelection,
            versionRight: rightSelection
          }
        }
      );
    }
  }, [props, selection]);

  const cellProps = (item, column) => {
    return onRowClick
      ? {
          style: { cursor: "pointer" },
          onClick: !column.hasActions ? () => onRowClick(item) : undefined,
          className: column.hasActions
            ? "euiTableCellContent--showOnHover euiTableCellContent--overflowingContent"
            : ""
        }
      : undefined;
  };

  const rowProps = item => ({
    className: isSelected(item) ? "euiTableRow-isSelected" : ""
  });

  const columns = [
    {
      field: "version",
      name: "Version",
      mobileOptions: {
        enlarge: true,
        fullWidth: true
      },
      sortable: true,
      width: "21%",
      render: version => (
        <EuiText
          size={defaultTextSize}
          style={{ display: "flex", alignItems: "center" }}>
          {version}&nbsp;&nbsp;
          {version === deployedVersion && (
            <EuiBadge color="default">Current</EuiBadge>
          )}
        </EuiText>
      )
    },
    {
      field: "status",
      name: "Status",
      width: "26%",
      render: status => <DeploymentStatusHealth status={status} />
    },
    {
      field: "created_at",
      name: "Created",
      width: "23%",
      render: date => <DateFromNow date={date} size={defaultTextSize} />
    },
    {
      field: "updated_at",
      name: "Updated",
      width: "23%",
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
      width: "110px",
      align: "right",
      hasActions: true,
      render: item => {
        const isItemSelected = isSelected(item);

        const actions = [
          {
            type: "icon",
            name: !isItemSelected ? "Compare versions" : "",
            description: !isItemSelected ? "Compare versions" : "",
            icon: isItemSelected ? "check" : "merge",
            available: () => true,
            enabled: () => isItemSelected || !readyForDiff,
            onClick: () =>
              isItemSelected
                ? unselectFromComparison(item)
                : selectForComparison(item)
          },
          ...(readyForDiff && isItemSelected
            ? [
                {
                  type: "button",
                  name: "Compare",
                  onClick: onShowDiff,
                  color: "primary",
                  size: "xs",
                  available: () => true,
                  enabled: () => true
                }
              ]
            : renderActions())
        ];

        return (
          <Fragment>
            {actions.reduce((actions, action, idx) => {
              const available = action.available
                ? action.available(item)
                : true;
              if (!available) {
                return actions;
              }

              const enabled = action.enabled ? action.enabled(item) : true;
              actions.push(
                <DefaultItemAction
                  key={idx}
                  action={{
                    ...action,
                    onClick: item => {
                      action.onClick(item);
                    }
                  }}
                  item={item}
                  enabled={enabled}
                  className="euiIconButton--action euiTableCellContent__hoverItem"
                />
              );

              return actions;
            }, [])}
          </Fragment>
        );
      }
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
      rowProps={rowProps}
      itemId="version"
      pagination={true}
      hasActions={true}
      sorting={{ sort: { field: "Version", direction: "desc" } }}
    />
  );
};

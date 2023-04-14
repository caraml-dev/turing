import React, { Fragment, useMemo } from "react";
import {
  EuiBadge,
  EuiBasicTable,
  EuiButtonEmpty,
  EuiCallOut,
  EuiSearchBar,
  EuiSpacer,
  EuiText,
} from "@elastic/eui";
import { useNavigate } from "react-router-dom";
import { useConfig } from "../../config";
import { EnsemblerType } from "../../services/ensembler/EnsemblerType";
import moment from "moment";
import { FormLabelWithToolTip } from "../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { DateFromNow } from "@caraml-dev/ui-lib";

export const ListEnsemblersTable = ({
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
  const searchQuery = useMemo(() => {
    const parts = [];
    if (!!filter.search) {
      parts.push(filter.search);
    }
    if (!!filter.type) {
      parts.push(`type:${filter.type}`);
    }

    return parts.join(" ");
  }, [filter]);

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
              <EuiBadge color="success">New</EuiBadge>
            </Fragment>
          )}
        </EuiText>
      ),
    },
    {
      field: "name",
      name: "Name",
      width: "30%",
      render: (name) => (
        <span className="eui-textTruncate" title={name}>
          {name}
        </span>
      ),
    },
    {
      field: "type",
      name: "Type",
      width: "15%",
    },
    {
      field: "created_at",
      name: "Created",
      sortable: true,
      width: "20%",
      render: (date) => <DateFromNow date={date} size={defaultTextSize} />,
    },
    {
      field: "updated_at",
      name: "Updated",
      width: "20%",
      render: (date) => <DateFromNow date={date} size={defaultTextSize} />,
    },
    {
      field: "id",
      name: (
        <FormLabelWithToolTip
          label="Actions"
          size={defaultIconSize}
          content="Ensembler actions"
        />
      ),
      align: "right",
      mobileOptions: {
        header: true,
        fullWidth: false,
      },
      width: "15%",
      render: (ensemblerId) => (
        <EuiButtonEmpty
          onClick={(_) => navigate(`../jobs?ensembler_id=${ensemblerId}`)}
          iconType="storage"
          size="xs">
          <EuiText size="xs">Batch Jobs</EuiText>
        </EuiButtonEmpty>
      ),
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
        field: "type",
        name: "Ensembler Type",
        multiSelect: false,
        options: EnsemblerType.values.map((type) => ({
          value: type.toString(),
          view: type.label,
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
      />
    </Fragment>
  );
};
